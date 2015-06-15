package gogstash

import (
	_ "github.com/tsaikd/gogstash/input/docker"
	_ "github.com/tsaikd/gogstash/input/dockerstats"
	_ "github.com/tsaikd/gogstash/input/exec"
	_ "github.com/tsaikd/gogstash/input/file"
	_ "github.com/tsaikd/gogstash/input/http"
	_ "github.com/tsaikd/gogstash/output/redis"
	_ "github.com/tsaikd/gogstash/output/report"
	_ "github.com/tsaikd/gogstash/output/stdout"
)
