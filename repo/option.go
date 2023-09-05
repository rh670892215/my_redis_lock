package repo

const (
	// 默认最大空闲连接数
	defaultMaxIdle = 10
	// 默认最大活跃连接数
	defaultMaxActive = 100
	// 默认空闲连接超时
	defaultIdleTimeOut = 10
)

type ClientOptions struct {
	// must
	netWork string
	address string
	appKey  string
	// optional
	maxIdle         int
	maxActive       int
	idleTimeout     int
	wait            bool
	maxConnLifeTime int
}

type ClientOption func(*ClientOptions)

func WithMaxIdle(maxIdle int) ClientOption {
	return func(c *ClientOptions) {
		c.maxActive = maxIdle
	}
}

func WithMaxActive(maxActive int) ClientOption {
	return func(c *ClientOptions) {
		c.maxActive = maxActive
	}
}

func WithIdleTimeout(idleTimeout int) ClientOption {
	return func(c *ClientOptions) {
		c.idleTimeout = idleTimeout
	}
}

func WithMaxConnLifeTime(maxConnLifeTime int) ClientOption {
	return func(c *ClientOptions) {
		c.maxConnLifeTime = maxConnLifeTime
	}
}

func WithWait(wait bool) ClientOption {
	return func(c *ClientOptions) {
		c.wait = wait
	}
}

func modifyClientOptions(c *ClientOptions) {
	if c.maxIdle <= 0 {
		c.maxIdle = defaultMaxIdle

	}

	if c.maxActive <= 0 {
		c.maxActive = defaultMaxActive
	}

	if c.idleTimeout <= 0 {
		c.idleTimeout = defaultIdleTimeOut
	}
}
