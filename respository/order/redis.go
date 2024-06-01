package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hanzalahimran7/MicroserviceInGo/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

type FindAllPage struct {
	Size   uint
	Offset uint
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("FAILED TO CREATE ORDER: %w", err)
	}
	key := fmt.Sprintf("order:%d", order.OrderID)
	tnx := r.Client.TxPipeline()
	res := tnx.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		tnx.Discard()
		return fmt.Errorf("FAILED TO SET %w", err)
	}
	if err := tnx.SAdd(ctx, "orders", key).Err(); err != nil {
		tnx.Discard()
		return fmt.Errorf("Failed to add order is sets")
	}
	if _, err := tnx.Exec(ctx); err != nil {
		return fmt.Errorf("Failed to exec")
	}
	return nil
}

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := fmt.Sprintf("order:%d", id)
	res, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, errors.New("order does not exist")
	} else if err != nil {
		return model.Order{}, fmt.Errorf("order get: %w", err)
	}
	var order model.Order
	if err := json.Unmarshal([]byte(res), &order); err != nil {
		return model.Order{}, fmt.Errorf("FAILED TO DECODE THE JSON: %w", err)
	}
	return order, nil
}

func (r *RedisRepo) DeleteById(ctx context.Context, id uint64) error {
	key := fmt.Sprintf("order:%d", id)
	tnx := r.Client.TxPipeline()
	err := tnx.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		tnx.Discard()
		return errors.New("order does not exist")
	} else if err != nil {
		tnx.Discard()
		return fmt.Errorf("order get: %w", err)
	}
	if err := tnx.SRem(ctx, "orders", key); err != nil {
		tnx.Discard()
		return fmt.Errorf("Failed to remove order from set")
	}
	if _, err := tnx.Exec(ctx); err != nil {
		return fmt.Errorf("Failed to exec")
	}
	return nil
}

func (r *RedisRepo) UpdateOrder(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("FAILED TO CREATE ORDER: %w", err)
	}
	key := fmt.Sprintf("order:%d", order.OrderID)
	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return errors.New("order does not exist")
	} else if err != nil {
		return fmt.Errorf("order get: %w", err)
	}
	return nil
}

func (r *RedisRepo) ListOrders(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", uint64(page.Offset), "*", int64(page.Size))
	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids %w", err)
	}
	if len(keys) == 0 {
		return FindResult{}, fmt.Errorf("NO ORDERS")
	}
	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids %w", err)
	}
	orders := make([]model.Order, len(xs))
	for i, x := range xs {
		x := x.(string)
		var order model.Order
		if err := json.Unmarshal([]byte(x), &order); err != nil {
			return FindResult{}, fmt.Errorf("failed to decode json %w", err)
		}
		orders[i] = order
	}
	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
