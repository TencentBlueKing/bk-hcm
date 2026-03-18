/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package cloudadaptor

import (
	"fmt"
	"sync"
	"time"

	"hcm/pkg/adaptor/aws"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

const refreshBeforeExpiry = 10 * time.Minute

// CachedCredential holds a cached STS temporary credential with its expiration.
type CachedCredential struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// CredentialCache is an in-process cache for STS AssumeRole temporary credentials.
type CredentialCache struct {
	mu    sync.Mutex
	cache map[string]*CachedCredential
}

// NewCredentialCache creates a new empty credential cache.
func NewCredentialCache() *CredentialCache {
	return &CredentialCache{
		cache: make(map[string]*CachedCredential),
	}
}

// GetOrRefresh returns cached credentials if valid, or calls STS to obtain new ones.
// cacheKey is constructed by the orchestration method (AwsWithAssumeRole) to support role chaining scenarios.
// externalID is optional; when non-empty it is passed to the STS AssumeRole call.
func (c *CredentialCache) GetOrRefresh(kt *kit.Kit, secret *types.BaseSecret, cacheKey, roleArn, sessionName, externalID string,
	site enumor.AccountSiteType) (*CachedCredential, error) {

	key := cacheKey
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	cached, exists := c.cache[key]

	// Cache hit, not near expiry — return directly.
	if exists && now.Before(cached.Expiration.Add(-refreshBeforeExpiry)) {
		return cached, nil
	}

	// Cache hit, near expiry — try refresh, fallback to old credential on failure.
	if exists && now.Before(cached.Expiration) {
		result, err := aws.AssumeRole(secret, roleArn, sessionName, externalID, site)
		if err != nil {
			logs.Warnf("refresh STS credential failed (using cached), key: %s, err: %v, rid: %s", key, err, kt.Rid)
			return cached, nil
		}
		refreshed := &CachedCredential{
			AccessKeyID:     result.AccessKeyID,
			SecretAccessKey: result.SecretAccessKey,
			SessionToken:    result.SessionToken,
			Expiration:      result.Expiration,
		}
		c.cache[key] = refreshed
		logs.Infof("STS credential refreshed, key: %s, expires: %v, rid: %s", key, refreshed.Expiration, kt.Rid)
		return refreshed, nil
	}

	// Cache miss or expired — must obtain new credentials.
	result, err := aws.AssumeRole(secret, roleArn, sessionName, externalID, site)
	if err != nil {
		return nil, fmt.Errorf("assume role [%s] failed: %w", key, err)
	}
	fresh := &CachedCredential{
		AccessKeyID:     result.AccessKeyID,
		SecretAccessKey: result.SecretAccessKey,
		SessionToken:    result.SessionToken,
		Expiration:      result.Expiration,
	}
	c.cache[key] = fresh
	logs.Infof("STS credential obtained, key: %s, expires: %v, rid: %s", key, fresh.Expiration, kt.Rid)
	return fresh, nil
}
