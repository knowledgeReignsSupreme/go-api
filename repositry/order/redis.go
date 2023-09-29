package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/knowledgeReignsSupreme/go-api.git/model"
	"github.com/redis/go-redis/v9"
)
type RedisRepo struct {
	Client *redis.Client
}

func OrderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert (ctx context.Context, order model.Order) error {
	data , err:= json.Marshal(order)
	
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}
	
	key := OrderIDKey(order.OrderID)
	txn := r.Client.TxPipeline()

	res := r.Client.SetNX(ctx, key, string(data), 0)

	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	if err := r.Client.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add to order set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %W", err)
	}

	return nil
}

var ErrNotExist = errors.New("order does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := OrderIDKey(id)

	value, err  := r.Client.Get(ctx,key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	}else if err != nil {
		return model.Order{}, fmt.Errorf("ger order: %w", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(value), &order)

	if err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order json: %w", err)
	}

	return order,nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := OrderIDKey(id)
	
	txn := r.Client.TxPipeline()
	
	 err  := r.Client.Del(ctx,key).Err()

	 if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	}else if err != nil {
		txn.Discard()
		return  fmt.Errorf("ger order: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %W", err)
	}

	return nil
}

func (r *RedisRepo) Update (ctx context.Context, order model.Order) error {
	data , err:= json.Marshal(order)

	
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}
	
	key := OrderIDKey(order.OrderID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	}else if err != nil {
		return  fmt.Errorf("ger order: %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size uint
	Offset uint
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}


func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", uint64(page.Offset), "*", int64(page.Size))
	 
	keys, cursor, err := res.Result()

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}
	
	if err != nil {
		return  FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()

	if err != nil {
		return  FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x := x.(string)
		var order model.Order

		err := json.Unmarshal([]byte(x), &order)

		if err != nil {
			return  FindResult{}, fmt.Errorf("failed to decode order: %w", err)
		}

		orders[i] = order
	}
	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}