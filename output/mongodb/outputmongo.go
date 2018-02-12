package outputmongodb

import (
	"context"
	"time"
	"sync"

	"github.com/tsaikd/KDGoLib/timeutil"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"gopkg.in/mgo.v2"
)

// ModuleName is the name used in config file
const ModuleName = "mongodb"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_output_mongodb_error"

// mgoPool used to create a pool of connections
type mgoPool struct {
	sync.Mutex
	pool            chan *mgo.Session
	poolCapacity	int
	firstSession	*mgo.Session
}

func (mpool *mgoPool) Get() *mgo.Session {
	mpool.Lock()
	defer mpool.Unlock()
	select {
	case session := <-mpool.pool:
		return session
	default:
		return mpool.firstSession.Copy()
	}
}

func (mpool *mgoPool) Put(session *mgo.Session) {
	mpool.Lock()
	defer mpool.Unlock()
	count := len(mpool.pool)
	if count >= mpool.poolCapacity {
		session.Close()
	} else {
		mpool.pool <- session
	}
}

// OutputConfig holds the configuration json fields and internal objects
type OutputConfig struct {
	config.OutputConfig
	Host              []string `json:"host"`
	Database          string   `json:"database,omitempty"`
	Collection        string   `json:"collection,omitempty"`
	Timeout           int      `json:"timeout,omitempty"`
	Connections       int      `json:"connections,omitempty"` // maximum number of socket connections
	Username          string   `json:"username"`
	Password          string   `json:"password"`
	Mechanism         string   `json:"mechanism,omitempty"`
	RetryInterval     int      `json:"retry_interval,omitempty"`
	RetryMax          int      `json:"retry_max,omitempty"`

	firstSession      *mgo.Session
	pool              *mgoPool
}

// DefaultOutputConfig returns an OutputConfig struct with default values
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		OutputConfig: config.OutputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Host:              []string{"localhost:27017"},
		Database:          "gogstash",
		Collection:        "allLogs",
		Timeout:           3,
		Connections:       10,
		Username:          "username",
		Password:          "password",
		Mechanism:         "MONGODB-CR",
		RetryInterval:     10,
		RetryMax:          -1,
	}
}

// errors
var (
	ErrorInsertMongoDBFailed1      = errutil.NewFactory("Insert MongoDB failed: %v")
	ErrorConnectionMongoDBFailed1  = errutil.NewFactory("Connection MongoDB failed: %v")
	ErrorMaxRetryMongoDBFailed1    = errutil.NewFactory("Max retry number must be greater than 1 or equal -1 for unlimited: %d")
)

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeOutputConfig, error) {
	var err error
	conf := DefaultOutputConfig()
	err = config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.RetryMax < -1 || conf.RetryMax == 0 {
		return nil, ErrorMaxRetryMongoDBFailed1.New(nil, conf.RetryMax)
	}

	// Init pool
	pool := &mgoPool{}
	pool.pool = make(chan *mgo.Session, conf.Connections)
	pool.poolCapacity = conf.Connections

	info := &mgo.DialInfo {
		Addrs:		conf.Host,
		Direct:		false,
		Timeout:	time.Duration(conf.Timeout) * time.Second,
		Database:	conf.Database,
		Username:	conf.Username,
		Password:	conf.Password,
		Mechanism:	conf.Mechanism,
	}

	conf.firstSession, err = mgo.DialWithInfo(info)
	if err != nil {
		return nil, ErrorConnectionMongoDBFailed1.New(err, info)
	}

	conf.pool = pool
	for i := 0; i < conf.Connections; i++ {
		pool.Put(conf.firstSession.Copy())
	}
	pool.firstSession = conf.firstSession

	return &conf, nil
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) error {
	var err error
	i := 0
	m := event.GetMap()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		session := t.pool.Get()
		coll := session.DB(t.Database).C(t.Collection)
		err = coll.Insert(m)
		t.pool.Put(session)
		if err == nil {
			return nil
		} else if t.RetryMax == -1 || i < t.RetryMax {
			i++
			timeout := time.Duration(t.RetryInterval) * time.Second
			timeutil.ContextSleep(ctx, timeout)
			continue
		}
		break
	}
	return ErrorInsertMongoDBFailed1.New(err, "Max retry")
}

