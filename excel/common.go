package excel

func NewConnecter() Connecter {
	return &connect{}
}

func UnmarshalXLSX(filePath string, container interface{}) error {
	conn := NewConnecter()
	err := conn.Open(filePath)
	if err != nil {
		return err
	}

	rd, err := conn.NewReader(container)
	if err != nil {
		conn.Close()
		return err
	}

	err = rd.ReadAll(container)
	if err != nil {
		conn.Close()
		rd.Close()
		return err
	}
	conn.Close()
	rd.Close()
	return nil
}
