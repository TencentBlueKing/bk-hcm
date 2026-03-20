## MODIFIED Requirements

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

#### Scenario: Service layer handles overwrite directly
- **GIVEN** the service struct holds a `client *client.ClientSet` field
- **WHEN** the handler receives an overwrite GPU demand order request
- **THEN** the handler SHALL call an internal service method to perform validation, deletion, and recreation, without going through logics.Controller

## REMOVED Requirements

### Requirement: Overwrite GPU demand order via logics Controller
**Reason**: GPU demand order overwrite is a standalone business flow that does not need cross-module shared logic. Moving it to the service layer aligns with the project's architecture convention.
**Migration**: The same logic is now implemented as a service method in `cmd/woa-server/service/plan/gpu_demand_order.go`. The handler calls `s.overwriteGpuDemandOrder(...)` directly instead of `s.planController.OverwriteGpuDemandOrder(...)`.
