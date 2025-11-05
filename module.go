package armcontrols

import (
	"context"
	"errors"
	"fmt"

	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	generic "go.viam.com/rdk/services/generic"
)

var (
	RepeatArmMovements = resource.NewModel("brad-grigsby", "arm-controls", "repeat-arm-movements")
)

func init() {
	resource.RegisterService(generic.API, RepeatArmMovements,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newArmControlsRepeatArmMovements,
		},
	)
}

type Config struct {
	Arm            string      `json:"arm"`
	JointPositions [][]float64 `json:"joint_positions"`
	NumRepeats     int         `json:"num_repeats"`
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.Arm == "" {
		return nil, nil, errors.New("arm must be specified and cannot be empty")
	}
	if len(cfg.JointPositions) == 0 {
		return nil, nil, errors.New("joint_positions must be specified and cannot be empty")
	}
	if cfg.NumRepeats <= 0 {
		return nil, nil, errors.New("num_repeats must be greater than zero")
	}

	return nil, nil, nil
}

type armControlsRepeatArmMovements struct {
	resource.AlwaysRebuild

	arm arm.Arm

	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()
}

func newArmControlsRepeatArmMovements(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (resource.Resource, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewRepeatArmMovements(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewRepeatArmMovements(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (resource.Resource, error) {
	arm, err := arm.FromProvider(deps, conf.Arm)
	if err != nil {
		return nil, fmt.Errorf("failed to get arm %q from provider: %w", conf.Arm, err)
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &armControlsRepeatArmMovements{
		arm:        arm,
		name:       name,
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *armControlsRepeatArmMovements) Name() resource.Name {
	return s.name
}

func (s *armControlsRepeatArmMovements) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	commandType, ok := cmd["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command must have a 'command' key of type string")
	}

	switch commandType {
	case "execute":
		for i := 0; i < s.cfg.NumRepeats; i++ {
			s.logger.Infof("Running iteration %v\n", i)
			for _, jp := range s.cfg.JointPositions {
				err := s.arm.MoveToJointPositions(s.cancelCtx, jp, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to move arm to joint positions on iteration %v: %w", i, err)
				}
			}
		}
		return map[string]any{"executed_repeats": s.cfg.NumRepeats}, nil
	case "move_to_index":
		index, ok := cmd["index"].(int)
		if !ok {
			return nil, fmt.Errorf("move_to_index command requires an 'index' key of type int")
		}
		if index < 0 || index >= len(s.cfg.JointPositions) {
			return nil, fmt.Errorf("index %d out of range", index)
		}
		jp := s.cfg.JointPositions[index]
		err := s.arm.MoveToJointPositions(s.cancelCtx, jp, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to move arm to joint positions at index %d: %w", index, err)
		}
		return map[string]any{"moved_to_index": index}, nil
	case "cancel":
		s.cancelFunc()
		s.cancelCtx, s.cancelFunc = context.WithCancel(context.Background())
		s.logger.Info("Cancelled current arm movement")
		return map[string]any{"cancelled": true}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", commandType)
	}
}

func (s *armControlsRepeatArmMovements) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
