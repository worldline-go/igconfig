package main

import (
	"context"
	"fmt"
	"sync"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"
)

func main() {
	consul := loader.Consul{}
	consul.Client, _ = loader.NewConsulFromEnv()

	updateHandler := func(input []byte) error {
		if input == nil {
			return nil
		}

		fmt.Println(string(input))

		return nil
	}

	// This context will handle cancellation for DynamicUpdate.
	// Meaning that when this context will be canceled - DynamicValue will also be stopped.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			err := consul.DynamicValue(ctx, loader.DynamicConfig{
				AppName:   "test",
				FieldName: "", // This can also be sub-key: 'struct/inner/field'
				Runner:    updateHandler,
			})

			fmt.Println(err.Error())
		}
	}()

	wg.Wait()
}
