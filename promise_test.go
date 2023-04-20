package promise

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPromiseAll(t *testing.T) {
	t.Run("all success", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := make([]Promise[int], 0, 10)
		for i := 0; i < 10; i++ {
			i := i
			promises = append(promises, NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(time.Duration(i*100) * time.Millisecond)
				resolve(i + 1)
			}))
		}
		results, err := All(ctx, promises)
		assert.NoError(t, err)
		assert.Equal(t, 10, len(results))
		assert.Equal(t, 1, results[0])
	})

	t.Run("return error if have a error", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := make([]Promise[int], 0, 10)
		caseErr := 5
		for i := 0; i < 10; i++ {
			i := i
			promises = append(promises, NewPromise[int](func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(time.Duration(i*100) * time.Millisecond)
				if i != caseErr {
					resolve(i + 1)
				} else {
					reject(fmt.Errorf("error at promise %d", i))
				}
			}))
		}
		results, err := All(ctx, promises)
		assert.Error(t, err)
		assert.Error(t, err, "error at promise 5")
		fmt.Println(results)
	})

	t.Run("return error if have a error 2", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := make([]Promise[int], 0, 100)
		caseErr := 4
		caseErr2 := 6
		for i := 0; i < 10; i++ {
			i := i
			promises = append(promises, NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(time.Duration(i*100) * time.Millisecond)
				if i != caseErr && i != caseErr2 {
					resolve(i + 1)
				} else {
					reject(fmt.Errorf("error at promise %d", i))
				}
			}))
		}
		results, err := All(ctx, promises)
		assert.Error(t, err)
		assert.True(t, "error at promise 4" == err.Error() || "error at promise 6" == err.Error())
		fmt.Println(results)
	})

	t.Run("all error", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(100 * time.Millisecond)
				reject(errors.New("error occur"))
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(200 * time.Millisecond)
				reject(errors.New("error occur"))
			}),
		}

		_, err := All(ctx, promises)
		assert.Error(t, err)
	})

	t.Run("all promise don't resolved or reject", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(100 * time.Millisecond)
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(200 * time.Millisecond)
			}),
		}

		results, err := All(ctx, promises)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(results))
	})

}

func TestPromiseAny(t *testing.T) {
	t.Run("return first case success", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(100 * time.Millisecond)
				resolve(1)
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(200 * time.Millisecond)
				resolve(2)
			}),
		}

		result, err := Any(ctx, promises)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("return second case success", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(100 * time.Millisecond)
				reject(errors.New("error occur"))
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(200 * time.Millisecond)
				resolve(2)
			}),
		}

		result, err := Any(ctx, promises)
		assert.NoError(t, err)
		assert.Equal(t, 2, result)
	})

	t.Run("return error if all error", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(100 * time.Millisecond)
				reject(errors.New("error occur"))
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(200 * time.Millisecond)
				reject(errors.New("error occur"))
			}),
		}

		_, err := Any(ctx, promises)
		assert.Error(t, err)
	})

	t.Run("return zero value of int (0) if all promise don't resolved or reject", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
			}),
		}

		result, err := Any[int](ctx, promises)
		assert.Nil(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("return zero value of pointer (nil) if all promise don't resolved or reject", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		type CustomStruct struct{}
		promises := []Promise[*CustomStruct]{
			NewPromise(func(ctx context.Context, resolve func(data *CustomStruct), reject func(err error)) {
			}),
			NewPromise(func(ctx context.Context, resolve func(data *CustomStruct), reject func(err error)) {
			}),
		}

		result, err := Any[*CustomStruct](ctx, promises)
		assert.Nil(t, err)
		assert.Nil(t, result)
	})
}

func TestPromiseRace(t *testing.T) {
	t.Run("return first case success", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(100 * time.Millisecond)
				resolve(1)
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(200 * time.Millisecond)
				resolve(2)
			}),
		}

		result, err := Race[int](ctx, promises)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("return first case error", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(100 * time.Millisecond)
				reject(errors.New("error occur 1"))
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(200 * time.Millisecond)
				reject(errors.New("error occur 2"))
			}),
		}

		result, err := Race(ctx, promises)
		assert.Error(t, err)
		assert.Equal(t, "error occur 1", err.Error())
		assert.Nil(t, result)
	})

	t.Run("return first case error", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(100 * time.Millisecond)
				reject(errors.New("error occur 1"))
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(200 * time.Millisecond)
				resolve(2)
			}),
		}

		result, err := Race(ctx, promises)
		assert.Error(t, err)
		assert.Equal(t, "error occur 1", err.Error())
		assert.Nil(t, result)
	})

	t.Run("return nil if both don't resolved, reject", func(t *testing.T) {
		t.Parallel()
		ctx := context.TODO()
		promises := []Promise[int]{
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(100 * time.Millisecond)
			}),
			NewPromise(func(ctx context.Context, resolve func(data int), reject func(err error)) {
				time.Sleep(200 * time.Millisecond)
			}),
		}

		result, err := Race(ctx, promises)
		assert.Nil(t, err)
		assert.Nil(t, result)
	})
}
