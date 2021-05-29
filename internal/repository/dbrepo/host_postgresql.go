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
	// get all services for host
	stmt := `SELECT
				hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit,
				hs.last_check, hs.status, hs.created_at, hs.updated_at,
				s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at
			FROM 
				host_services hs
				left join services s on (s.id = hs.service_id)
			WHERE 
				host_id = $1
			`

	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil {
		return h, err
	}
	defer rows.Close()

	var hostServices []models.HostService

	for rows.Next() {
		var hs models.HostService
		err := rows.Scan(
			&hs.ID,
			&hs.HostID,
			&hs.ServiceID,
			&hs.Active,
			&hs.ScheduleNumber,
			&hs.ScheduleUnit,
			&hs.LastCheck,
			&hs.Status,
			&hs.CreatedAt,
			&hs.UpdatedAt,
			&hs.Service.ID,
			&hs.Service.ServiceName,
			&hs.Service.Active,
			&hs.Service.Icon,
			&hs.Service.CreatedAt,
			&hs.Service.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
		}
		hostServices = append(hostServices, hs)
	}
	if err := rows.Err(); err != nil {
		return h, err
	}
	h.HostServices = hostServices

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
		stmt := `SELECT
		hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit,
		hs.last_check, hs.status, hs.created_at, hs.updated_at,
		s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at
	FROM 
		host_services hs
		left join services s on (s.id = hs.service_id)
	WHERE 
		host_id = $1
	`
		serviceRows, err := m.DB.QueryContext(ctx, stmt, h.ID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer serviceRows.Close()

		var hostServices []models.HostService
		for serviceRows.Next() {
			var hs models.HostService
			err := serviceRows.Scan(
				&hs.ID,
				&hs.HostID,
				&hs.ServiceID,
				&hs.Active,
				&hs.ScheduleNumber,
				&hs.ScheduleUnit,
				&hs.LastCheck,
				&hs.Status,
				&hs.CreatedAt,
				&hs.UpdatedAt,
				&hs.Service.ID,
				&hs.Service.ServiceName,
				&hs.Service.Active,
				&hs.Service.Icon,
				&hs.Service.CreatedAt,
				&hs.Service.UpdatedAt,
			)
			if err != nil {
				return nil, err
			}
			hostServices = append(hostServices, hs)
		}
		if err := serviceRows.Err(); err != nil {
			return nil, err
		}
		h.HostServices = hostServices
		hosts = append(hosts, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}

// Updates the active status of a host service
func (m *postgresDBRepo) UpdateHostServiceStatus(hostID, serviceID, active int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE host_services
			SET
				active = $1
			WHERE
				host_id = $2 and service_id = $3
			`
	if _, err := m.DB.ExecContext(ctx, stmt, active, hostID, serviceID); err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetAllServiceStatusCounts() (pending int, healthy int, warning int, problem int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := ` select
				(select count(id) from host_services where active = 1 and status = 'pending') as pending,
				(select count(id) from host_services where active = 1 and status = 'healthy') as healthy,
				(select count(id) from host_services where active = 1 and status = 'warning') as warning,
				(select count(id) from host_services where active = 1 and status = 'problem') as problem
			`
	row := m.DB.QueryRowContext(ctx, stmt)
	err = row.Scan(&pending, &healthy, &warning, &problem)
	if err != nil {
		return
	}

	return
}

func (m *postgresDBRepo) GetServicesByStatus(status string) ([]models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT
		hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit,
		hs.last_check, hs.status, hs.created_at, hs.updated_at,
		h.host_name, s.service_name 
	FROM 
		host_services hs
		left join hosts h on (h.id = hs.host_id)
		left join services s on (s.id = hs.service_id)
	WHERE 
		status = $1
		and hs.active = 1
	ORDER BY
		host_name, service_name
	`

	var hostServices []models.HostService

	rows, err := m.DB.QueryContext(ctx, stmt, status)
	if err != nil {
		return hostServices, err
	}
	defer rows.Close()

	for rows.Next() {
		var hs models.HostService
		if err := rows.Scan(
			&hs.ID,
			&hs.HostID,
			&hs.ServiceID,
			&hs.Active,
			&hs.ScheduleNumber,
			&hs.ScheduleUnit,
			&hs.LastCheck,
			&hs.Status,
			&hs.CreatedAt,
			&hs.UpdatedAt,
			&hs.HostName,
			&hs.Service.ServiceName,
		); err != nil {
			return nil, err
		}
		hostServices = append(hostServices, hs)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hostServices, nil
}

func (m *postgresDBRepo) GetHostServiceByID(id int) (models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT
				hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit,
				hs.last_check, hs.status, hs.created_at, hs.updated_at,
				s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at
			FROM 
				host_services hs
			LEFT JOIN
				services s on (hs.service_id = s.id)
			WHERE 
				hs.id = $1`

	var hs models.HostService
	row := m.DB.QueryRowContext(ctx, query, id)
	if err := row.Scan(
		&hs.ID,
		&hs.HostID,
		&hs.ServiceID,
		&hs.Active,
		&hs.ScheduleNumber,
		&hs.ScheduleUnit,
		&hs.LastCheck,
		&hs.Status,
		&hs.CreatedAt,
		&hs.UpdatedAt,
		&hs.Service.ID,
		&hs.Service.ServiceName,
		&hs.Service.Active,
		&hs.Service.Icon,
		&hs.Service.CreatedAt,
		&hs.Service.UpdatedAt,
	); err != nil {
		return hs, err
	}

	return hs, nil

}

// Updates a host service
func (m *postgresDBRepo) UpdateHostService(hs models.HostService) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE host_services
			SET
				host_id = $1, service_id = $2, active = $3,
				schedule_number = $4, schedule_unit = $5,
				last_check = $6, status = $7, updated_at = $8
			WHERE
				id = $9
			`
	if _, err := m.DB.ExecContext(ctx, stmt,
		hs.HostID,
		hs.ServiceID,
		hs.Active,
		hs.ScheduleNumber,
		hs.ScheduleUnit,
		hs.LastCheck,
		hs.Status,
		hs.UpdatedAt,
		hs.ID,
	); err != nil {
		return err
	}

	return nil
}
