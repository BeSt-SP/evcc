package meter

import (
	"errors"
	"fmt"

	"github.com/evcc-io/evcc/provider"
)

// BuildMeasurements returns typical meter measurement getters from config
func BuildMeasurements(
	power *provider.Config,
	energy *provider.Config,
	currents []provider.Config,
	voltages []provider.Config,
	powers []provider.Config,
) (
	func() (float64, error),
	func() (float64, error),
	func() (float64, float64, float64, error),
	func() (float64, float64, float64, error),
	func() (float64, float64, float64, error),
	error,
) {
	var powerG func() (float64, error)
	if power != nil {
		var err error
		powerG, err = provider.NewFloatGetterFromConfig(*power)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("power: %w", err)
		}
	}

	var energyG func() (float64, error)
	if energy != nil {
		var err error
		energyG, err = provider.NewFloatGetterFromConfig(*energy)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("energy: %w", err)
		}
	}

	currentsG, err := BuildPhaseProviders(currents)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("currents: %w", err)
	}

	voltagesG, err := BuildPhaseProviders(voltages)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("voltages: %w", err)
	}

	powersG, err := BuildPhaseProviders(powers)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("powers: %w", err)
	}

	return powerG, energyG, currentsG, voltagesG, powersG, nil
}

// BuildPhaseProviders returns phases getter for given config
func BuildPhaseProviders(providers []provider.Config) (func() (float64, float64, float64, error), error) {
	if len(providers) == 0 {
		return nil, nil
	}

	if len(providers) != 3 {
		return nil, errors.New("need one per phase, total three")
	}

	var phases [3]func() (float64, error)
	for idx, prov := range providers {
		c, err := provider.NewFloatGetterFromConfig(prov)
		if err != nil {
			return nil, fmt.Errorf("[%d] %w", idx, err)
		}

		phases[idx] = c
	}

	return CollectPhaseProviders(phases), nil
}

// CollectPhaseProviders combines phase getters into combined api function
func CollectPhaseProviders(g [3]func() (float64, error)) func() (float64, float64, float64, error) {
	return func() (float64, float64, float64, error) {
		var res [3]float64
		for idx, currentG := range g {
			c, err := currentG()
			if err != nil {
				return 0, 0, 0, err
			}

			res[idx] = c
		}

		return res[0], res[1], res[2], nil
	}
}