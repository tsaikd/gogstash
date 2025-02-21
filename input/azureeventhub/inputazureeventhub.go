package inputazureeventhub

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "azureeventhub"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig
	EventHubNamespaceConnectionString string         `json:"eventhub_namespace_connection_string"`
	EventHub                          string         `json:"eventhub"`
	StorageConnectionString           string         `json:"storage_connection_string"`
	StorageContainer                  string         `json:"storage_container"`
	ConsumerGroup                     string         `json:"group"`
	OffsetEarliest                    bool           `json:"offset_earliest"`
	Extras                            map[string]any `json:"extras"`
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		ConsumerGroup: azeventhubs.DefaultConsumerGroup,
	}
}

func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	conf.Codec, err = config.GetCodec(ctx, config.ConfigRaw{"type": "azureeventhubjson"}, "azureeventhubjson")
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func (t *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error) {
	// Create checkpoint store and consumer client

	containerClient, err := container.NewClientFromConnectionString(t.StorageConnectionString, t.StorageContainer, nil)
	if err != nil {
		goglog.Logger.Errorf("Error instanciating storage client from connection string: %v", err)
		return err
	}

	checkpointStore, err := checkpoints.NewBlobStore(containerClient, nil)
	if err != nil {
		goglog.Logger.Errorf("Error instanciating checkpoint store: %v", err)
		return err
	}

	consumerClient, err := azeventhubs.NewConsumerClientFromConnectionString(t.EventHubNamespaceConnectionString, t.EventHub, t.ConsumerGroup, nil)
	if err != nil {
		goglog.Logger.Errorf("Error instanciating consumer client from connection string: %v", err)
		return err
	}

	defer consumerClient.Close(context.TODO())

	// Create a processor to load balance eventhub partitiond

	var earliest bool = t.OffsetEarliest
	var latest bool = !t.OffsetEarliest

	processor, err := azeventhubs.NewProcessor(consumerClient, checkpointStore, &azeventhubs.ProcessorOptions{
		LoadBalancingStrategy: azeventhubs.ProcessorStrategyBalanced,
		StartPositions: azeventhubs.StartPositions{
			Default: azeventhubs.StartPosition{
				Earliest: &earliest,
				Latest:   &latest,
			},
		},
	})

	if err != nil {
		goglog.Logger.Errorf("Error instanciating processor: %v", err)
		return err
	}

	wg := &sync.WaitGroup{}

	go func() {
		maxLoopWithoutEventBeforeReconnect := 5

		for {
			partitionClient := processor.NextPartitionClient(context.TODO())
			if partitionClient == nil {
				break
			}

			wg.Add(1)

			go func() {
				defer partitionClient.Close(context.TODO())
				defer wg.Done()

				loopWithoutEvent := 0

				goglog.Logger.Infof("New client initialized %s/%s", t.EventHub, partitionClient.PartitionID())

				for {
					receiveCtx, receiveCtxCancel := context.WithTimeout(context.TODO(), time.Minute)
					events, err := partitionClient.ReceiveEvents(receiveCtx, 100, nil)
					receiveCtxCancel()

					for {
						if ctx.Err() == context.Canceled {
							break
						}

						receiveCtx, receiveCtxCancel := context.WithTimeout(ctx, time.Minute)
						events, err := partitionClient.ReceiveEvents(receiveCtx, 100, nil)
						receiveCtxCancel()

						if err != nil && !errors.Is(err, context.DeadlineExceeded) {
							if ctx.Err() != context.Canceled {
								if eventHubError := (*azeventhubs.Error)(nil); errors.As(err, &eventHubError) && eventHubError.Code == azeventhubs.ErrorCodeOwnershipLost {
									goglog.Logger.Debugln("Partition rebalanced")
								} else {
									goglog.Logger.Errorf("Error while processing partition: %v", err)
								}
							}
							break
						}

						for _, event := range events {
							ok, err := t.Codec.Decode(ctx, []byte(event.Body), t.Extras, []string{}, msgChan)
							if !ok {
								goglog.Logger.Errorf("Error while decoding message to msg chan: %v", err)
							}
						}

						// it's possible to get zero events if the partition is empty, or if no new events have arrived
						// since your last receive.
						if len(events) != 0 {
							// Update the checkpoint with the last event received. If we lose ownership of this partition or
							// have to restart the next owner will start from this point.
							if err := partitionClient.UpdateCheckpoint(ctx, events[len(events)-1], &azeventhubs.UpdateCheckpointOptions{}); err != nil {
								if ctx.Err() != context.Canceled {
									goglog.Logger.Warnf("Error during checkpoints update: %v", err)
									checkpointLastUpdateFailed = true
								}
							} else {
								if checkpointLastUpdateFailed {
									checkpointLastUpdateFailed = false
									goglog.Logger.Warnf("checkpoints updated")
								} else {
									goglog.Logger.Debugln("checkpoints updated")
								}
							}
						}
					}

					if len(events) == 0 {
						loopWithoutEvent++
						if loopWithoutEvent > maxLoopWithoutEventBeforeReconnect {
							goglog.Logger.Infof("MaxLoopWithoutEventBeforeReconnect %s/%s", t.EventHub, partitionClient.PartitionID())
							return
						}

						continue
					}

					loopWithoutEvent = 0

					for _, event := range events {
						ok, err := t.Codec.Decode(context.TODO(), []byte(event.Body), t.Extras, []string{}, msgChan)
						if err != nil && !ok {
							goglog.Logger.Errorf("Error while decoding message to msg chan: %v", err)
						}
					}

					// Update the checkpoint with the last event received. If we lose ownership of this partition or
					// have to restart the next owner will start from this point.
					if err := partitionClient.UpdateCheckpoint(context.TODO(), events[len(events)-1]); err != nil {
						goglog.Logger.Warnf("Error during checkpoints update: %v for %s/%s", err, t.EventHub, partitionClient.PartitionID())
						return
					}
				}
			}()
		}
	}()

	processorCtx, processorCtxCancel := context.WithCancel(context.TODO())
	defer processorCtxCancel()

	if err := processor.Run(processorCtx); err != nil {
		goglog.Logger.Errorf("Error, unable to start processor: %v", err)
		return err
	}

	goglog.Logger.Debugln("Waiting completion of all processors...")

	wg.Wait()
	goglog.Logger.Debugln("Terminated")

	return nil
}
