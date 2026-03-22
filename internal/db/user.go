package db

import "database/sql"

func (s *Store) GetAdminUser() (*User, error) {
	row := s.DB.QueryRow(`SELECT id, name, passwd, roles FROM sys_users WHERE name='admin' LIMIT 1`)
	u := &User{}
	err := row.Scan(&u.ID, &u.Name, &u.Passwd, &u.Roles)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) CreateAdminUser(hashedPassword string) error {
	_, err := s.DB.Exec(`INSERT INTO sys_users (name, passwd, roles) VALUES ('admin', ?, 'admin')`, hashedPassword)
	return err
}

func (s *Store) UpdateAdminPassword(hashedPassword string) error {
	_, err := s.DB.Exec(`UPDATE sys_users SET passwd=? WHERE name='admin'`, hashedPassword)
	return err
}
