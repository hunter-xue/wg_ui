package db

import "database/sql"

func (s *Store) GetServer() (*Server, error) {
	row := s.DB.QueryRow(`SELECT id, name, address, listen_port, private_key, public_key, mtu,
		COALESCE(dns,''), COALESCE(post_up,''), COALESCE(post_down,''),
		COALESCE(endpoint,''), COALESCE(comments,'')
		FROM wg_server LIMIT 1`)

	srv := &Server{}
	err := row.Scan(&srv.ID, &srv.Name, &srv.Address, &srv.ListenPort,
		&srv.PrivateKey, &srv.PublicKey, &srv.MTU,
		&srv.DNS, &srv.PostUp, &srv.PostDown, &srv.Endpoint, &srv.Comments)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func (s *Store) CreateServer(srv *Server) error {
	result, err := s.DB.Exec(`INSERT INTO wg_server
		(name, address, listen_port, private_key, public_key, mtu, dns, post_up, post_down, endpoint, comments)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		srv.Name, srv.Address, srv.ListenPort, srv.PrivateKey, srv.PublicKey, srv.MTU,
		srv.DNS, srv.PostUp, srv.PostDown, srv.Endpoint, srv.Comments)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	srv.ID = id
	return nil
}

func (s *Store) UpdateServer(srv *Server) error {
	_, err := s.DB.Exec(`UPDATE wg_server SET
		name=?, address=?, listen_port=?, private_key=?, public_key=?, mtu=?,
		dns=?, post_up=?, post_down=?, endpoint=?, comments=?
		WHERE id=?`,
		srv.Name, srv.Address, srv.ListenPort, srv.PrivateKey, srv.PublicKey, srv.MTU,
		srv.DNS, srv.PostUp, srv.PostDown, srv.Endpoint, srv.Comments, srv.ID)
	return err
}

func (s *Store) DeleteServer(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM wg_server WHERE id=?`, id)
	return err
}
