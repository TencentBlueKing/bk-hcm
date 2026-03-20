# Overwrite GPU Demand Order

## Purpose

This capability provides an endpoint to overwrite an existing GPU demand order by replacing all its sub-orders with a new set. It enforces status preconditions, extension field validation, and business-level access control.

## Requirements

### Requirement: Overwrite GPU demand order with new sub-orders
The system SHALL provide a PATCH endpoint at `/api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/order/overwrite` that accepts an order_id and a new details list, validates preconditions, deletes existing sub-orders, and creates new sub-orders under the same main order. The implementation SHALL reside in the `cmd/woa-server/service/plan/` package as a method on the service struct, not on the logics Controller.

#### Scenario: Successful overwrite of a fully rejected order
- **GIVEN** a GPU demand main order with status `REJECT_ALL` and all sub-orders in status `REJECT` or `TERMINATE`
- **WHEN** the user submits a valid overwrite request with order_id and new details
- **THEN** the system SHALL delete all existing sub-orders under this order, create new sub-orders from the provided details with status `INIT`, reset the main order status to `INIT`, and return success

#### Scenario: Overwrite rejected when main order is not fully rejected
- **GIVEN** a GPU demand main order with status other than `REJECT_ALL` and `INIT`
- **WHEN** the user submits an overwrite request for this order
- **THEN** the system SHALL return an error indicating the order status does not allow overwrite

#### Scenario: Overwrite rejected when sub-orders have non-terminal status
- **GIVEN** a GPU demand main order with status `REJECT_ALL` but at least one sub-order in status `INIT`, `PENDING`, or `DONE`
- **WHEN** the user submits an overwrite request for this order
- **THEN** the system SHALL return an error indicating not all sub-orders are in rejected or terminated status

#### Scenario: Service layer handles overwrite directly
- **GIVEN** the service struct holds a `client *client.ClientSet` field
- **WHEN** the handler receives an overwrite GPU demand order request
- **THEN** the handler SHALL call an internal service method to perform validation, deletion, and recreation, without going through logics.Controller

### Requirement: Extension validation on overwrite details
The system SHALL validate extension fields in the new details against the template schema before performing the overwrite, using the same validation logic as order creation.

#### Scenario: Extension validation passes
- **GIVEN** all details have valid extension fields matching the template schema
- **WHEN** the overwrite request is processed
- **THEN** the system SHALL proceed with deleting old sub-orders and creating new ones

#### Scenario: Extension validation fails
- **GIVEN** at least one detail has an invalid extension field (wrong type, missing required value, unknown demand_type)
- **WHEN** the overwrite request is processed
- **THEN** the system SHALL return a detailed validation error and NOT modify any existing data

### Requirement: Business authorization on overwrite
The system SHALL enforce business-level access control, ensuring the user has access to the specified bk_biz_id and the order belongs to that business.

#### Scenario: User has no access to the business
- **GIVEN** a user without permission to the specified bk_biz_id
- **WHEN** the user submits an overwrite request
- **THEN** the system SHALL return a permission denied error

#### Scenario: Order does not belong to the specified business
- **GIVEN** an order that belongs to bk_biz_id=100 but the request targets bk_biz_id=200
- **WHEN** the user submits an overwrite request
- **THEN** the system SHALL return an error indicating the order does not belong to the business
