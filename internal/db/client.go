package db

import "database/sql"

func (s *Store) ListClients(serverID int64, orderBy string) ([]Client, error) {
	if orderBy == "" {
		orderBy = "name"
	}
	// only allow safe column names
	switch orderBy {
	case "name", "address", "id":
	default:
		orderBy = "name"
	}

	rows, err := s.DB.Query(`SELECT id, server_id, name, address, COALESCE(listen_port,0),
		private_key, public_key, allow_ips, mtu, COALESCE(dns,''),
		COALESCE(description,''), COALESCE(comments,''), disabled, persistent_keepalive
		FROM wg_clients WHERE server_id=? ORDER BY `+orderBy, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var c Client
		if err := rows.Scan(&c.ID, &c.ServerID, &c.Name, &c.Address, &c.ListenPort,
			&c.PrivateKey, &c.PublicKey, &c.AllowIPs, &c.MTU, &c.DNS,
			&c.Description, &c.Comments, &c.Disabled, &c.PersistentKeepalive); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

func (s *Store) GetClient(id int64) (*Client, error) {
	row := s.DB.QueryRow(`SELECT id, server_id, name, address, COALESCE(listen_port,0),
		private_key, public_key, allow_ips, mtu, COALESCE(dns,''),
		COALESCE(description,''), COALESCE(comments,''), disabled, persistent_keepalive
		FROM wg_clients WHERE id=?`, id)

	c := &Client{}
	err := row.Scan(&c.ID, &c.ServerID, &c.Name, &c.Address, &c.ListenPort,
		&c.PrivateKey, &c.PublicKey, &c.AllowIPs, &c.MTU, &c.DNS,
		&c.Description, &c.Comments, &c.Disabled, &c.PersistentKeepalive)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) CreateClient(c *Client) error {
	result, err := s.DB.Exec(`INSERT INTO wg_clients
		(server_id, name, address, listen_port, private_key, public_key, allow_ips, mtu,
		 dns, description, comments, disabled, persistent_keepalive)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ServerID, c.Name, c.Address, c.ListenPort, c.PrivateKey, c.PublicKey,
		c.AllowIPs, c.MTU, c.DNS, c.Description, c.Comments, c.Disabled, c.PersistentKeepalive)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = id
	return nil
}

func (s *Store) UpdateClient(c *Client) error {
	_, err := s.DB.Exec(`UPDATE wg_clients SET
		server_id=?, name=?, address=?, listen_port=?, private_key=?, public_key=?,
		allow_ips=?, mtu=?, dns=?, description=?, comments=?, disabled=?, persistent_keepalive=?
		WHERE id=?`,
		c.ServerID, c.Name, c.Address, c.ListenPort, c.PrivateKey, c.PublicKey,
		c.AllowIPs, c.MTU, c.DNS, c.Description, c.Comments, c.Disabled, c.PersistentKeepalive, c.ID)
	return err
}

func (s *Store) SetClientDisabled(id int64, disabled bool) error {
	d := 0
	if disabled {
		d = 1
	}
	_, err := s.DB.Exec(`UPDATE wg_clients SET disabled=? WHERE id=?`, d, id)
	return err
}
