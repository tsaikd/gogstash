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
		Timeout:           10,
		Connections:       10,
		Username:          "username",
		Password:          "password",
		Mechanism:         "MONGODB-CR",
		RetryInterval:     10,
	}
}

// errors
var (
	ErrorInsertMongoDBFailed1  = errutil.NewFactory("insert MongoDB failed: %v")
)

// InitHandler initialize the output plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeOutputConfig, error) {
	var err error
	conf := DefaultOutputConfig()
	err = config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	conf.pool = pool
	for i := 0; i < conf.Connections; i++ {
		pool.Put(conf.firstSession.Copy())
	}
	pool.firstSession = conf.firstSession

	return &conf, nil
}

// Output event
func (t *OutputConfig) Output(ctx context.Context, event logevent.LogEvent) (err error) {
	m := event.GetMap()

	// try to log forever
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		session := t.pool.Get()
		coll := session.DB(t.Database).C(t.Collection)
		err := coll.Insert(m)
		if err == nil {
			t.pool.Put(session)
			return nil
		}
		ErrorInsertMongoDBFailed1.New(err, event)
		t.pool.Put(session)

		timeout := time.Duration(t.RetryInterval) * time.Second
		timeutil.ContextSleep(ctx, timeout)
	}
}

