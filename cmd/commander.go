package cmd

type Commander struct {
	Root       *Root
	Get        *Get
	GetDevices *GetDevices
	GetPods    *GetPods
}

func NewCommander() (*Commander, error) {
	root := NewRoot()
	get := NewGet(root.Command)
	getDevices := NewGetDevices(get.Command)
	getPods := NewGetPods(get.Command)

	if err := root.Command.Execute(); err != nil {
		return nil, err
	}

	return &Commander{
		Root:       root,
		Get:        get,
		GetDevices: getDevices,
		GetPods:    getPods,
	}, nil
}
