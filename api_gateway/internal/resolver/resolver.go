package resolver

import (
	"api_gateway/internal/graph/model"

	orderpb "api_gateway/internal/proto/orderpb"
	trackingpb "api_gateway/internal/proto/trackingpb"
	userpb "api_gateway/internal/proto/userpb"
)

// Resolver holds all the gRPC clients connecting to the downstream microservices.
type Resolver struct {
	UserClient     userpb.UserServiceClient
	OrderClient    orderpb.OrderServiceClient
	TrackingClient trackingpb.TrackingServiceClient
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func protoToUser(u *userpb.User) *model.User {
	if u == nil {
		return nil
	}
	return &model.User{ID: u.GetId(), Name: u.GetName(), Email: u.GetEmail()}
}

func protoToOrder(o *orderpb.Order) *model.Order {
	if o == nil {
		return nil
	}
	return &model.Order{
		ID:          o.GetId(),
		UserID:      o.GetUserId(),
		ProductName: o.GetProductName(),
		Quantity:    int(o.GetQuantity()),
		Status:      o.GetStatus(),
	}
}

func protoToTracking(t *trackingpb.OrderTracking) *model.OrderTracking {
	if t == nil {
		return nil
	}
	return &model.OrderTracking{
		ID:       t.GetId(),
		OrderID:  t.GetOrderId(),
		Status:   t.GetStatus(),
		Location: t.GetLocation(),
	}
}
