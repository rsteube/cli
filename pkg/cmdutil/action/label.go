package action

import (
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

type label struct {
	Name        string
	Description string
}

type labelsQuery struct {
	Data struct {
		Repository struct {
			Labels struct {
				Nodes []label
			}
		}
	}
}

func ActionLabels(cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(args []string) carapace.Action {
		var queryResult labelsQuery
		return GraphQlAction(cmd, `repository(owner: $owner, name: $repo){ labels(first: 100) { nodes { name, description } } }`, &queryResult, func() carapace.Action {
			labels := queryResult.Data.Repository.Labels.Nodes
			vals := make([]string, len(labels)*2)
			for index, label := range labels {
				vals[index*2] = label.Name
				vals[index*2+1] = label.Description
			}
			return carapace.ActionValuesDescribed(vals...)
		})
	})
}

func ActionMultiLabels(cmd *cobra.Command) carapace.Action {
	return carapace.ActionMultiParts(",", func(args, parts []string) carapace.Action {
		return ActionLabels(cmd).Invoke(args).Filter(parts).ToA()
	})
}
