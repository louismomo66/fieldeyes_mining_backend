package cache

import "fmt"

// InvalidateUserAnalytics removes all cached analytics for a given user.
// Call this whenever income or expense records are created, updated, or deleted.
func (c *Client) InvalidateUserAnalytics(userID uint) {
	if !c.enabled {
		return
	}
	c.DeletePattern(fmt.Sprintf("analytics:*:%d:*", userID))
	c.Delete(
		fmt.Sprintf("analytics:summary:%d", userID),
		fmt.Sprintf("analytics:breakdown:%d", userID),
	)
}
