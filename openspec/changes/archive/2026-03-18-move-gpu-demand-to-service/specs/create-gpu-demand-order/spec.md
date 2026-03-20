## MODIFIED Requirements

### Requirement: Logics interface extension
The `plan.Logics` interface SHALL NOT include `CreateGpuDemandOrder` method. The GPU demand order creation logic SHALL be implemented as a method on the service struct in `cmd/woa-server/service/plan/` package, directly calling data-service via `client.DataService()`.

#### Scenario: Interface method removed from Logics
- **GIVEN** the plan.Logics interface definition
- **WHEN** code references plan.Logics
- **THEN** the interface SHALL NOT contain `CreateGpuDemandOrder` method

#### Scenario: Service layer handles order creation directly
- **GIVEN** the service struct holds a `client *client.ClientSet` field
- **WHEN** the handler receives a create GPU demand order request
- **THEN** the handler SHALL call an internal service method to create the main order and sub-orders, without going through logics.Controller

## REMOVED Requirements

### Requirement: Create GPU demand order via logics Controller
**Reason**: GPU demand order creation is a standalone business flow that does not need cross-module shared logic. Moving it to the service layer aligns with the project's architecture convention.
**Migration**: The same logic is now implemented as a service method in `cmd/woa-server/service/plan/gpu_demand_order.go`. All callers (handler functions in the same package) call the service method directly instead of `s.planController.CreateGpuDemandOrder(...)`.
