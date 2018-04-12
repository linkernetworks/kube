package metricsclient

import (
	client "github.com/influxdata/influxdb/client/v2"
)

// rawQuery convenience function to query the database
func rawQuery(c client.Client, db, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Database: db,
		Command:  cmd,
	}
	c.Query(q)
	if response, err := c.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
