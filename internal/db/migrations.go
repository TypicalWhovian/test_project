package db

import (
	"database/sql"
	"github.com/GuiaBolso/darwin"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Initial migration",
			Script:      `
						create extension if not exists "uuid-ossp";
						
						create or replace function set_updated_at() returns trigger
						  language plpgsql
						as
						$$
						begin
						  new.updated_at = NOW();
						  return new;
						end;
						$$;
						
						create table if not exists tasks
						(
							id              uuid                     default uuid_generate_v4() not null
								constraint tasks_pk primary key,
							request_id      uuid                                                not null,
							steps_completed integer                  default 0                  not null
								constraint is_number_of_steps_correct
									check (steps_completed >= 0 and steps_completed <= 10),
							created_at      timestamp with time zone default now()              not null,
							updated_at      timestamp with time zone default now()              not null
						);`,
		},
	}
)

func Migrate(db *sql.DB) error {
	driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})
	d := darwin.New(driver, migrations, nil)
	return d.Migrate()
}
