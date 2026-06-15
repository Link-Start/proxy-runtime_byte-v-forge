package mihomo

import "fmt"

func configProjectionStageError(stage string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("mihomo config projection %s failed: %w", stage, err)
}
