package outputcond

import (
	"context"

	"github.com/Knetic/govaluate"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	filtercond "github.com/tsaikd/gogstash/filter/cond"
	"golang.org/x/sync/errgroup"
)

// ModuleName is the name used in config file
const ModuleName = "cond"

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig

	Condition     string             `json:"condition"`   // condition need to test
	OutputRaw     []config.ConfigRaw `json:"output"`      // filters when satisfy the condition
	ElseOutputRaw []config.ConfigRaw `json:"else_output"` // filters when does not met the condition
	outputs       []config.TypeOutputConfig
	elseOutputs   []config.TypeOutputConfig
	expression    *govaluate.EvaluableExpression
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}
}

// InitHandler initialize the output plugin
func InitHandler(
	ctx context.Context,
	raw config.ConfigRaw,
	control config.Control,
) (config.TypeOutputConfig, error) {
	conf := DefaultOutputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}
	if conf.Condition == "" {
		goglog.Logger.Warn("output cond config condition empty, ignored")
		return &conf, nil
	}
	conf.outputs, err = config.GetOutputs(ctx, conf.OutputRaw, control)
	if err != nil {
		return nil, err
	}
	if len(conf.outputs) <= 0 {
		goglog.Logger.Warn("output cond config outputs empty, ignored")
		return &conf, nil
	}
	if len(conf.ElseOutputRaw) > 0 {
		conf.elseOutputs, err = config.GetOutputs(ctx, conf.ElseOutputRaw, control)
		if err != nil {
			return nil, err
		}
	}
	conf.expression, err = govaluate.NewEvaluableExpressionWithFunctions(conf.Condition, filtercond.BuiltInFunctions)
	return &conf, err
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
	if t.expression != nil {
		ep := filtercond.EventParameters{Event: &event}
		ret, err := t.expression.Eval(&ep)
		if err != nil {
			return err
		}
		if r, ok := ret.(bool); ok {
			if r {
				eg, ctx2 := errgroup.WithContext(ctx)
				for _, output := range t.outputs {
					func(output config.TypeOutputConfig) {
						eg.Go(func() error {
							if err2 := output.Output(ctx2, event); err2 != nil {
								goglog.Logger.Errorf("output module %q failed: %v\n", output.GetType(), err2)
							}
							return nil
						})
					}(output)
				}
				return eg.Wait()
			} else if len(t.elseOutputs) > 0 {
				eg, ctx2 := errgroup.WithContext(ctx)
				for _, output := range t.elseOutputs {
					func(output config.TypeOutputConfig) {
						eg.Go(func() error {
							if err2 := output.Output(ctx2, event); err2 != nil {
								goglog.Logger.Errorf("output module %q failed: %v\n", output.GetType(), err2)
							}
							return nil
						})
					}(output)
				}
				return eg.Wait()
			}
		} else {
			goglog.Logger.Warn("output cond condition returns not a boolean, ignored")
		}
	}
	return nil
}
