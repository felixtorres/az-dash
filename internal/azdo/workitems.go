package azdo

import (
	"fmt"
	"net/url"
	"strings"
)

// QueryWorkItems runs a WIQL query and returns full work item data.
// $top is sent as a query parameter to limit results from the WIQL endpoint.
func (c *Client) QueryWorkItems(org, project, wiql string, top int) ([]WorkItem, error) {
	apiURL := fmt.Sprintf("%s/wit/wiql", c.projectURL(org, project))
	if top > 0 {
		apiURL += fmt.Sprintf("?$top=%d&api-version=%s", top, apiVersion)
	} else {
		apiURL += "?api-version=" + apiVersion
	}

	body := map[string]interface{}{
		"query": wiql,
	}

	var result WIQLResult
	if err := c.postRaw(apiURL, body, &result); err != nil {
		return nil, err
	}

	if len(result.WorkItems) == 0 {
		return nil, nil
	}

	return c.getWorkItemsByIDs(org, project, result.WorkItems)
}

func (c *Client) getWorkItemsByIDs(org, project string, refs []WIQLWorkItemRef) ([]WorkItem, error) {
	const batchSize = 200

	var all []WorkItem
	for i := 0; i < len(refs); i += batchSize {
		end := i + batchSize
		if end > len(refs) {
			end = len(refs)
		}

		ids := make([]string, end-i)
		for j, ref := range refs[i:end] {
			ids[j] = fmt.Sprintf("%d", ref.ID)
		}

		apiURL := fmt.Sprintf("%s/wit/workitems", c.orgURL(org))
		params := url.Values{
			"ids":     {strings.Join(ids, ",")},
			"$expand": {"all"},
		}

		var resp ListResponse[WorkItem]
		if err := c.get(apiURL, params, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Value...)
	}

	return all, nil
}
