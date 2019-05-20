package shims

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func parseOrderTOMLs(orders *[]order, orderFilesDir string) error {
	orderFiles, err := ioutil.ReadDir(orderFilesDir)
	if err != nil {
		return err
	}

	for _, file := range orderFiles {
		orderTOML, err := parseOrderTOML(filepath.Join(orderFilesDir, file.Name()))
		if err != nil {
			return err
		}

		*orders = append(*orders, orderTOML)
	}

	return nil
}

func parseOrderTOML(path string) (order, error) {
	var order order
	if _, err := toml.DecodeFile(path, &order); err != nil {
		return order, err
	}
	return order, nil
}

func combineOrders(orders []order) order {
	finalOrder := initOrder()

	for i := len(orders) - 1; i >= 0; i-- {
		var combinedGroups []group
		for _, currGroup := range orders[i].Groups {
			combinedGroups = prependCurrentGroup(finalOrder.Groups, currGroup, combinedGroups)
		}

		finalOrder.Groups = combinedGroups
	}

	return finalOrder
}

func prependCurrentGroup(groupsSoFar []group, currGroup group, combinedGroups []group) []group {
	for _, soFar := range groupsSoFar {
		labels := append(currGroup.Labels, soFar.Labels...)
		buildpacks := append(currGroup.Buildpacks, soFar.Buildpacks...)
		filterDuplicateBuildpacks(&buildpacks)

		combinedGroup := group{
			Labels:     labels,
			Buildpacks: buildpacks,
		}
		combinedGroups = append(combinedGroups, combinedGroup)
	}
	return combinedGroups
}

func filterDuplicateBuildpacks(bps *[]buildpack) {
	for i := range *bps {
		for j := i + 1; j < len(*bps); {
			if (*bps)[i].ID == (*bps)[j].ID {
				*bps = append((*bps)[:j], (*bps)[j+1:]...)
			} else {
				j++
			}
		}
	}
}

func encodeTOML(dest string, data interface{}) error {
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	return toml.NewEncoder(destFile).Encode(data)
}

func initOrder() order {
	return order{Groups: []group{{
		Labels:     []string{},
		Buildpacks: []buildpack{},
	}}}
}
