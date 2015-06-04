package gopencils

import "fmt"

func (resource *Resource) ProcessedError() error {
	if resource.Raw == nil {
		return fmt.Errorf("Error: empty response")
	}

	if resource.Raw.StatusCode < 200 && resource.Raw.StatusCode >= 300 {
		return fmt.Errorf("Error(%s) %s %s", resource.Raw.Status, resource.Raw.Request.Method, resource.Url)
	} else {
		return nil
	}

}
