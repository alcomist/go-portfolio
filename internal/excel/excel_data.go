package excel

import "strings"

type Row map[string]string

func newRow() Row {
	return make(map[string]string)
}

type Data struct {
	data    []Row
	current int
}

func NewData(rows [][]string) *Data {

	e := &Data{}
	e.current = -1

	data := make([]Row, 0)

	columns := rows[0]

	for _, row := range rows[1:] {

		r := newRow()

		for i, v := range row {

			if i >= len(columns) {
				continue
			}

			column := columns[i]
			r[strings.ToLower(strings.TrimSpace(column))] = strings.TrimSpace(v)
		}

		data = append(data, r)
	}

	e.data = data
	return e
}

func (d *Data) Len() int {
	return len(d.data)
}

func (d *Data) HasNext() bool {
	return d.current < len(d.data)-1
}

func (d *Data) Reset() {
	d.current = -1
}

func (d *Data) Next() {
	d.current++
}

func (d *Data) Row() Row {
	return d.data[d.current]
}
