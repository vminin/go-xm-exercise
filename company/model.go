package company

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Company struct {
	ID      int     `json:"id"`
	Name    *string `json:"name"`
	Code    *string `json:"code"`
	Country *string `json:"country"`
	Website *string `json:"website"`
	Phone   *string `json:"phone"`
}

type ErrNotFound struct {
	msg string
}

func (e ErrNotFound) Error() string {
	return e.msg
}

type Model struct {
	db *pgxpool.Pool
}

func NewModel(db *pgxpool.Pool) Model {
	return Model{db}
}

func (m Model) All(form url.Values) ([]Company, error) {
	sql, args, err := filteringQuery("select id, name, code, country, website, phone from company", form)
	if err != nil {
		return nil, ErrNotFound{err.Error()}
	}

	rows, err := m.db.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		c := Company{}
		err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.Country, &c.Website, &c.Phone)
		if err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}

	if rows.Err() != nil {
		return nil, err
	}

	if len(companies) == 0 {
		return nil, ErrNotFound{"companies not found"}
	}

	return companies, nil
}

func filteringQuery(query string, form url.Values) (string, []any, error) {
	var (
		i    int
		args []any
		b    strings.Builder

		attrset = map[string]bool{
			"name":    true,
			"code":    true,
			"country": true,
			"website": true,
			"phone":   true,
		}
	)

	b.WriteString(query)
	b.WriteString(" where 1=1")

	for k, val := range form {
		if !attrset[strings.ToLower(k)] {
			return "", nil, fmt.Errorf("unknown attribute %q", k)
		}

		i++
		fmt.Fprintf(&b, " and %v=$%v", k, i)
		args = append(args, val[0])
	}

	return b.String(), args, nil
}

func (m Model) ByID(id string) (*Company, error) {
	sql := "select id, name, code, country, website, phone from company where id = $1"
	c := Company{}
	err := m.db.QueryRow(context.Background(), sql, id).Scan(&c.ID, &c.Name, &c.Code, &c.Country, &c.Website, &c.Phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrNotFound{fmt.Sprintf("company with id %q not found", id)}
		}
		return nil, err
	}
	return &c, nil
}

func (m Model) New(c Company) (*Company, error) {
	tx, err := m.db.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	sql := "insert into company (name, code, country, website, phone) values ($1, $2, $3, $4, $5) returning id"
	err = tx.QueryRow(context.Background(), sql, c.Name, c.Code, c.Country, c.Website, c.Phone).Scan(&c.ID)
	if err != nil {
		return nil, err
	}

	tx.Commit(context.Background())
	return &c, nil
}

func (m Model) Delete(c Company) error {
	tx, err := m.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	sql := "delete from company where id = $1"
	tag, err := m.db.Exec(context.Background(), sql, c.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("failed to delete company with id %v", c.ID)
	}

	tx.Commit(context.Background())
	return nil
}

func (m Model) Update(c Company) error {
	tx, err := m.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	sql := `update company 
		set name = $1,
			code = $2,
			country = $3,
			website = $4,
			phone = $5
		where id = $6`
	tag, err := m.db.Exec(context.Background(), sql, c.Name, c.Code, c.Country, c.Website, c.Phone, c.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("failed to update company with id %v", c.ID)
	}

	tx.Commit(context.Background())
	return nil
}
