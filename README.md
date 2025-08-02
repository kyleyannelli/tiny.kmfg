# tiny.kmfg

Administrative url aliasing service for managing long and large url collections.

# Setup

Run the binary or build it yourself with `go build`

#### Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| `KMFG_TINY_API_PORT` | 30108 | port for the API server |
| `KMFG_TINY_WEB_PORT` | 30109 | port for the web management server |
| `KMFG_TINY_DB` | `./tiny.kmfg.db` | sqlite database location |
| `KMFG_TINY_DB_LOG` | `false` | set to true for database logging |

### Run the Service
![Screenshot of tiny.kmfg startup logs](/screenshots/log_example.png)
