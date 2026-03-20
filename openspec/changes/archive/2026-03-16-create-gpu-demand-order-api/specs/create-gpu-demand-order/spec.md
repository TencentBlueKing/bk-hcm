## ADDED Requirements

### Requirement: Create GPU demand order API endpoint
The system SHALL provide a POST endpoint at `/api/v1/woa/bizs/{bk_biz_id}/plans/resource/gpu/order/create` that creates a GPU demand order (main order) and its associated sub-orders from the details provided by the frontend.

#### Scenario: Successful order creation
- **WHEN** a valid request is sent with `op_product_id`, `op_product_name`, and a `details` array containing valid sub-order data
- **THEN** the system SHALL create one main order record in `res_plan_demand_gpu_order` table with status `INIT`, and create one sub-order record per detail in `res_plan_demand_gpu_suborder` table with status `INIT`, and return the main order ID

#### Scenario: Empty details array
- **WHEN** a request is sent with an empty `details` array
- **THEN** the system SHALL return an `InvalidParameter` error

#### Scenario: Details exceed maximum limit
- **WHEN** a request is sent with more than 100 items in the `details` array
- **THEN** the system SHALL return an `InvalidParameter` error

#### Scenario: No GPU template configured
- **WHEN** no GPU demand template exists in the system
- **THEN** the system SHALL return an error indicating that GPU template must be configured first

#### Scenario: Sub-order creation failure after main order created
- **WHEN** the main order is created successfully but sub-order batch creation fails
- **THEN** the system SHALL return the error from sub-order creation without rolling back the main order

### Requirement: Request type definition
The system SHALL define a `CreateGpuDemandOrderReq` type in `pkg/api/woa-server/` with fields: `op_product_id` (int64, required), `op_product_name` (string, required), and `details` (array, required, min=1, max=100).

#### Scenario: Valid request structure
- **WHEN** a request contains all required fields with correct types
- **THEN** the `Validate()` method SHALL return nil

#### Scenario: Missing required fields
- **WHEN** a request is missing `op_product_id`, `op_product_name`, or `details`
- **THEN** the `Validate()` method SHALL return a validation error

### Requirement: Detail to sub-order conversion
Each detail item SHALL be converted to a sub-order with the following mapping: `demand_type` maps to `DemandType`, `demand_year` to `DemandYear`, `demand_month` to `DemandMonth`, `gpu_num` to `GPUNum`, `qpm_max` (float64) truncated to `QpmMax` (int64), `extension` (array) JSON-serialized to `Extension` (JsonField). Each sub-order SHALL be associated with the main order via `order_id`.

#### Scenario: Extension field serialization
- **WHEN** a detail has `extension: ["scene1", "H20", 12]`
- **THEN** the sub-order `Extension` field SHALL contain the JSON serialization `["scene1","H20",12]`

#### Scenario: QPM float to int conversion
- **WHEN** a detail has `qpm_max: 13.7`
- **THEN** the sub-order `QpmMax` field SHALL be `13` (truncated to int64)

### Requirement: Authorization check
The service handler SHALL verify that the requesting user has `meta.ResPlan` + `meta.Create` permission for the specified `bk_biz_id` before proceeding with order creation.

#### Scenario: Unauthorized user
- **WHEN** a user without ResPlan Create permission sends a create request
- **THEN** the system SHALL return an authorization error

### Requirement: Logics interface extension
The `plan.Logics` interface SHALL include a `CreateGpuDemandOrder` method that accepts `kit.Kit`, `bkBizID int64`, and the request type, returning `(string, error)` where string is the main order ID.

#### Scenario: Interface method availability
- **WHEN** the plan controller is initialized
- **THEN** the `CreateGpuDemandOrder` method SHALL be callable through the `Logics` interface
