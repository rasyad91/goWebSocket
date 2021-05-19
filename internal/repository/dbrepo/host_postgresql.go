package dbrepo

import (
	"context"
	"log"
	"time"
	"vigilate/internal/models"
)

// InsertHost inserts a host into the database
func (m *postgresDBRepo) InsertHost(h models.Host) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var newID int

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return newID, nil
	}

	query := `insert into hosts 
				(host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at)
			values
				($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			returning id
			`
	if err := tx.QueryRowContext(ctx, query,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID); err != nil {
		log.Println(tx.Rollback())
		return newID, err
	}

	stmt := ` insert into host_services (host_id, service_id, active, schedule_number, schedule_unit,
				last_check, created_at, updated_at, status)
			values ($1, 1, 0, 3, 'm', $2,$3, $4, 'pending')
	`

	if _, err := tx.ExecContext(ctx, stmt, newID, time.Now(), time.Now(), time.Now()); err != nil {
		log.Println(tx.Rollback())
		return newID, err
	}
	if err := tx.Commit(); err != nil {
		log.Println(err)
	}

	return newID, nil
}

func (m *postgresDBRepo) GetHostByID(id int) (models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := ` select id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at
				from hosts 
				where id = $1`

	var h models.Host

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&h.ID,
		&h.HostName,
		&h.CanonicalName,
		&h.URL,
		&h.IP,
		&h.IPV6,
		&h.Location,
		&h.OS,
		&h.Active,
		&h.CreatedAt,
		&h.UpdatedAt,
	)
	if err != nil {
		return h, err
	}
	return h, nil
}

func (m *postgresDBRepo) UpdateHost(h models.Host) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE hosts 
			SET
				host_name = $1,
				canonical_name = $2,
				url = $3,
				ip = $4,
				ipv6 = $5,
				location = $6,
				os = $7,
				active = $8,
				updated_at = $9
			WHERE
				id = $10
				`
	_, err := m.DB.ExecContext(ctx, query,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		h.ID,
	)
	if err != nil {
		return err
	}
	return nil

}

func (m *postgresDBRepo) GetAllHosts() ([]models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select 
				id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at
			from 
				hosts
			order by 
				id`

	hosts := []models.Host{}

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var h models.Host
		err := rows.Scan(
			&h.ID,
			&h.HostName,
			&h.CanonicalName,
			&h.URL,
			&h.IP,
			&h.IPV6,
			&h.Location,
			&h.OS,
			&h.Active,
			&h.CreatedAt,
			&h.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}
