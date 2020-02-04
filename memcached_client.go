package gomemcached

type MemcachedClient struct {
	cluster *Cluster
}

func NewMemcachedClient(addrs []string, maxConnPerServer uint32) Client {
	m := &MemcachedClient{}
	m.cluster = createCluster(addrs, maxConnPerServer)
	return m
}

func (m *MemcachedClient) AddServer(addr string, maxConnPerServer uint32) error {
	return m.cluster.AddServer2Cluster(addr, maxConnPerServer)
}

func (m *MemcachedClient) SetServerErrorCallback(errCall ServerErrorCallback) {
	m.cluster.serverErrCallback = errCall
}

func (m *MemcachedClient) Exit() {
	m.cluster.exit()
}

func (m *MemcachedClient) exec(key string, cmdFunc func(cmder *Commander) error) error {
	var err error
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			m.cluster.ReleaseServerCommand(server, cmder)
		} else if _, ok := err.(*StatusError); ok {
			m.cluster.ReleaseServerCommand(server, cmder)
		}
	}()

	err = cmdFunc(cmder)
	return err
}

func (m *MemcachedClient) Get(key string, value interface{}) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.get(key, value)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Set(args *KeyArgs) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		args.useMsgpack = true
		modifyCAS, resErr = cmder.store(OPCODE_SET, args)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) SetRawData(args *KeyArgs) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		args.useMsgpack = false
		modifyCAS, resErr = cmder.store(OPCODE_SET, args)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Add(args *KeyArgs) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		args.useMsgpack = true
		modifyCAS, resErr = cmder.store(OPCODE_ADD, args)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) AddRawData(args *KeyArgs) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		args.useMsgpack = false
		modifyCAS, resErr = cmder.store(OPCODE_ADD, args)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Replace(args *KeyArgs) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		args.useMsgpack = true
		modifyCAS, resErr = cmder.store(OPCODE_REPLACE, args)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) ReplaceRawData(args *KeyArgs) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		args.useMsgpack = false
		modifyCAS, resErr = cmder.store(OPCODE_REPLACE, args)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Append(args *KeyArgs) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.append(OPCODE_APPEND, args)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Prepend(args *KeyArgs) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.append(OPCODE_PREPEND, args)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Increment(args *KeyArgs) (uint64, uint64, error) {
	var value uint64
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		value, modifyCAS, resErr = cmder.atomic(OPCODE_INCR, args)
		return resErr
	})

	return value, modifyCAS, err
}

func (m *MemcachedClient) Decrement(args *KeyArgs) (uint64, uint64, error) {
	var value uint64
	var modifyCAS uint64
	var resErr error

	err := m.exec(args.Key, func(cmder *Commander) error {
		value, modifyCAS, resErr = cmder.atomic(OPCODE_DECR, args)
		return resErr
	})

	return value, modifyCAS, err
}

func (m *MemcachedClient) TouchAtomicValue(key string) (uint64, error) {
	var value uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		value, resErr = cmder.touchAtomicValue(key)
		return resErr
	})

	return value, err
}
