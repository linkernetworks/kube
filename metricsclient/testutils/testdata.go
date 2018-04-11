package testutils

var testdata = []byte(`
[
  {
    "Cmd": "SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"namespace_name\"",
    "Query": {
      "Command": "SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"namespace_name\"",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "uptime",
              "columns": [
                "key",
                "value"
              ],
              "values": [
                [
                  "namespace_name",
                  "default"
                ],
                [
                  "namespace_name",
                  "docker"
                ],
                [
                  "namespace_name",
                  "kube-system"
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"nodename\"",
    "Query": {
      "Command": "SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"nodename\"",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "uptime",
              "columns": [
                "key",
                "value"
              ],
              "values": [
                [
                  "nodename",
                  "docker-for-desktop"
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"cpu/usage_rate\" WHERE \"type\" = 'node' AND \"nodename\" = 'docker-for-desktop' ORDER BY DESC LIMIT 10",
    "Query": {
      "Command": "SELECT value FROM \"cpu/usage_rate\" WHERE \"type\" = 'node' AND \"nodename\" = 'docker-for-desktop' ORDER BY DESC LIMIT 10",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "cpu/usage_rate",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  683
                ],
                [
                  "2018-04-11T03:51:00Z",
                  601
                ],
                [
                  "2018-04-11T03:50:00Z",
                  662
                ],
                [
                  "2018-04-11T03:49:00Z",
                  641
                ],
                [
                  "2018-04-11T03:48:00Z",
                  682
                ],
                [
                  "2018-04-11T03:47:00Z",
                  792
                ],
                [
                  "2018-04-11T03:46:00Z",
                  625
                ],
                [
                  "2018-04-11T03:45:00Z",
                  636
                ],
                [
                  "2018-04-11T03:44:00Z",
                  639
                ],
                [
                  "2018-04-11T03:43:00Z",
                  692
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"cpu/usage_rate\" WHERE \"type\" = 'node' AND \"nodename\" = 'docker-for-desktop' ORDER BY DESC LIMIT 1",
    "Query": {
      "Command": "SELECT value FROM \"cpu/usage_rate\" WHERE \"type\" = 'node' AND \"nodename\" = 'docker-for-desktop' ORDER BY DESC LIMIT 1",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "cpu/usage_rate",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  683
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"memory/usage\" WHERE \"type\" = 'node' AND \"nodename\" = 'docker-for-desktop' ORDER BY DESC LIMIT 10",
    "Query": {
      "Command": "SELECT value FROM \"memory/usage\" WHERE \"type\" = 'node' AND \"nodename\" = 'docker-for-desktop' ORDER BY DESC LIMIT 10",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "memory/usage",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  3020943360
                ],
                [
                  "2018-04-11T03:51:00Z",
                  3009150976
                ],
                [
                  "2018-04-11T03:50:00Z",
                  3006140416
                ],
                [
                  "2018-04-11T03:49:00Z",
                  3003420672
                ],
                [
                  "2018-04-11T03:48:00Z",
                  3006529536
                ],
                [
                  "2018-04-11T03:47:00Z",
                  3005272064
                ],
                [
                  "2018-04-11T03:46:00Z",
                  3000446976
                ],
                [
                  "2018-04-11T03:45:00Z",
                  2996092928
                ],
                [
                  "2018-04-11T03:44:00Z",
                  3003559936
                ],
                [
                  "2018-04-11T03:43:00Z",
                  2994634752
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"memory/usage\" WHERE \"type\" = 'node' AND \"nodename\" = 'docker-for-desktop' ORDER BY DESC LIMIT 1",
    "Query": {
      "Command": "SELECT value FROM \"memory/usage\" WHERE \"type\" = 'node' AND \"nodename\" = 'docker-for-desktop' ORDER BY DESC LIMIT 1",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "memory/usage",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  3020943360
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"pod_name\" WHERE \"namespace_name\"='default'",
    "Query": {
      "Command": "SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"pod_name\" WHERE \"namespace_name\"='default'",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "uptime",
              "columns": [
                "key",
                "value"
              ],
              "values": [
                [
                  "pod_name",
                  "cv-server-55d4669796-jrbbt"
                ],
                [
                  "pod_name",
                  "face-recognition-worker-deployment-9ff4c85b6-tvj5w"
                ],
                [
                  "pod_name",
                  "fileserver-5ac3321cf8efa90001f512c6-fs"
                ],
                [
                  "pod_name",
                  "fileserver-5ac3338ff8efa90001f512c7-fs"
                ],
                [
                  "pod_name",
                  "fileserver-5ac3342ef8efa90001f512c8-fs"
                ],
                [
                  "pod_name",
                  "fileserver-5acc20d76ea2a500015f6f8d-fs"
                ],
                [
                  "pod_name",
                  "gearman-job-server-76db5cf944-2c2nx"
                ],
                [
                  "pod_name",
                  "influxdb-0"
                ],
                [
                  "pod_name",
                  "job-5acc20f9acbff200079782bb-5-run-1-9xg4f"
                ],
                [
                  "pod_name",
                  "job-5acc51ecacbff200079782bc-5-run-1-ftzn8"
                ],
                [
                  "pod_name",
                  "job-5acc8a15acbff200079782bd-5-run-1-4m26f"
                ],
                [
                  "pod_name",
                  "jobserver-766b96bf7f-wjg4l"
                ],
                [
                  "pod_name",
                  "jobupdater-687c6795c7-vgnzd"
                ],
                [
                  "pod_name",
                  "kube-registry-proxy-kscf5"
                ],
                [
                  "pod_name",
                  "kudis-79bb8fbc57-grvc6"
                ],
                [
                  "pod_name",
                  "mongo-0"
                ],
                [
                  "pod_name",
                  "nodesync-ffdd578b9-r592c"
                ],
                [
                  "pod_name",
                  "notebook-5ac3321cf8efa90001f512c6-5cc43761"
                ],
                [
                  "pod_name",
                  "notebook-5ac3338ff8efa90001f512c7-5cc43761"
                ],
                [
                  "pod_name",
                  "notebook-5ac3342ef8efa90001f512c8-5cc43761"
                ],
                [
                  "pod_name",
                  "notebook-5acc20d76ea2a500015f6f8d-5cc43761"
                ],
                [
                  "pod_name",
                  "redis-6fd4fb8c74-4kvdj"
                ],
                [
                  "pod_name",
                  "registry-0"
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = 'default' AND \"type\" = 'pod' AND \"pod_name\" = 'mongo-0' ORDER BY DESC LIMIT 10",
    "Query": {
      "Command": "SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = 'default' AND \"type\" = 'pod' AND \"pod_name\" = 'mongo-0' ORDER BY DESC LIMIT 10",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "cpu/usage_rate",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  18
                ],
                [
                  "2018-04-11T03:51:00Z",
                  20
                ],
                [
                  "2018-04-11T03:50:00Z",
                  20
                ],
                [
                  "2018-04-11T03:49:00Z",
                  20
                ],
                [
                  "2018-04-11T03:48:00Z",
                  19
                ],
                [
                  "2018-04-11T03:47:00Z",
                  20
                ],
                [
                  "2018-04-11T03:46:00Z",
                  20
                ],
                [
                  "2018-04-11T03:45:00Z",
                  20
                ],
                [
                  "2018-04-11T03:44:00Z",
                  19
                ],
                [
                  "2018-04-11T03:43:00Z",
                  20
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = 'default' AND \"type\" = 'pod' AND \"pod_name\" = 'mongo-0' ORDER BY DESC LIMIT 1",
    "Query": {
      "Command": "SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = 'default' AND \"type\" = 'pod' AND \"pod_name\" = 'mongo-0' ORDER BY DESC LIMIT 1",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "cpu/usage_rate",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  18
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = 'default' AND \"type\" = 'pod' AND \"pod_name\" = 'mongo-0' ORDER BY DESC LIMIT 10",
    "Query": {
      "Command": "SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = 'default' AND \"type\" = 'pod' AND \"pod_name\" = 'mongo-0' ORDER BY DESC LIMIT 10",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "memory/usage",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  406003712
                ],
                [
                  "2018-04-11T03:51:00Z",
                  402804736
                ],
                [
                  "2018-04-11T03:50:00Z",
                  401428480
                ],
                [
                  "2018-04-11T03:49:00Z",
                  398487552
                ],
                [
                  "2018-04-11T03:48:00Z",
                  403582976
                ],
                [
                  "2018-04-11T03:47:00Z",
                  402096128
                ],
                [
                  "2018-04-11T03:46:00Z",
                  399007744
                ],
                [
                  "2018-04-11T03:45:00Z",
                  394584064
                ],
                [
                  "2018-04-11T03:44:00Z",
                  401481728
                ],
                [
                  "2018-04-11T03:43:00Z",
                  401887232
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = 'default' AND \"type\" = 'pod' AND \"pod_name\" = 'mongo-0' ORDER BY DESC LIMIT 1",
    "Query": {
      "Command": "SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = 'default' AND \"type\" = 'pod' AND \"pod_name\" = 'mongo-0' ORDER BY DESC LIMIT 1",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "memory/usage",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  406003712
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"container_name\" WHERE \"namespace_name\" = 'default' AND \"pod_name\" = 'mongo-0'",
    "Query": {
      "Command": "SHOW TAG VALUES FROM \"uptime\" WITH KEY = \"container_name\" WHERE \"namespace_name\" = 'default' AND \"pod_name\" = 'mongo-0'",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "uptime",
              "columns": [
                "key",
                "value"
              ],
              "values": [
                [
                  "container_name",
                  "mongo"
                ],
                [
                  "container_name",
                  "mongo-sidecar"
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = 'default' AND \"pod_name\" = 'mongo-0' AND \"container_name\" = 'mongo-sidecar' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT 10",
    "Query": {
      "Command": "SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = 'default' AND \"pod_name\" = 'mongo-0' AND \"container_name\" = 'mongo-sidecar' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT 10",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "cpu/usage_rate",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  3
                ],
                [
                  "2018-04-11T03:51:00Z",
                  3
                ],
                [
                  "2018-04-11T03:50:00Z",
                  4
                ],
                [
                  "2018-04-11T03:49:00Z",
                  4
                ],
                [
                  "2018-04-11T03:48:00Z",
                  3
                ],
                [
                  "2018-04-11T03:47:00Z",
                  4
                ],
                [
                  "2018-04-11T03:46:00Z",
                  4
                ],
                [
                  "2018-04-11T03:45:00Z",
                  4
                ],
                [
                  "2018-04-11T03:44:00Z",
                  3
                ],
                [
                  "2018-04-11T03:43:00Z",
                  3
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = 'default' AND \"pod_name\" = 'mongo-0' AND \"container_name\" = 'mongo-sidecar' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT 1",
    "Query": {
      "Command": "SELECT value FROM \"cpu/usage_rate\" WHERE \"cluster_name\" = 'default' AND \"pod_name\" = 'mongo-0' AND \"container_name\" = 'mongo-sidecar' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT 1",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "cpu/usage_rate",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  3
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = 'default' AND \"pod_name\" = 'mongo-0' AND \"container_name\" = 'mongo-sidecar' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT 10",
    "Query": {
      "Command": "SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = 'default' AND \"pod_name\" = 'mongo-0' AND \"container_name\" = 'mongo-sidecar' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT 10",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "memory/usage",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  57970688
                ],
                [
                  "2018-04-11T03:51:00Z",
                  54681600
                ],
                [
                  "2018-04-11T03:50:00Z",
                  53399552
                ],
                [
                  "2018-04-11T03:49:00Z",
                  50401280
                ],
                [
                  "2018-04-11T03:48:00Z",
                  55455744
                ],
                [
                  "2018-04-11T03:47:00Z",
                  54099968
                ],
                [
                  "2018-04-11T03:46:00Z",
                  51101696
                ],
                [
                  "2018-04-11T03:45:00Z",
                  46718976
                ],
                [
                  "2018-04-11T03:44:00Z",
                  53403648
                ],
                [
                  "2018-04-11T03:43:00Z",
                  53993472
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = 'default' AND \"pod_name\" = 'mongo-0' AND \"container_name\" = 'mongo-sidecar' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT 1",
    "Query": {
      "Command": "SELECT value FROM \"memory/usage\" WHERE \"cluster_name\" = 'default' AND \"pod_name\" = 'mongo-0' AND \"container_name\" = 'mongo-sidecar' AND \"type\" = 'pod_container' ORDER BY DESC LIMIT 1",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "memory/usage",
              "columns": [
                "time",
                "value"
              ],
              "values": [
                [
                  "2018-04-11T03:52:00Z",
                  57970688
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  },
  {
    "Cmd": "SHOW STATS",
    "Query": {
      "Command": "SHOW STATS",
      "Database": "k8s",
      "Precision": "",
      "Chunked": false,
      "ChunkSize": 0,
      "Parameters": null
    },
    "Resp": {
      "Results": [
        {
          "Series": [
            {
              "name": "runtime",
              "columns": [
                "Alloc",
                "Frees",
                "HeapAlloc",
                "HeapIdle",
                "HeapInUse",
                "HeapObjects",
                "HeapReleased",
                "HeapSys",
                "Lookups",
                "Mallocs",
                "NumGC",
                "NumGoroutine",
                "PauseTotalNs",
                "Sys",
                "TotalAlloc"
              ],
              "values": [
                [
                  42636920,
                  42241297,
                  42636920,
                  44679168,
                  45170688,
                  594772,
                  24920064,
                  89849856,
                  10005,
                  42836069,
                  523,
                  38,
                  1254053100,
                  100251896,
                  8618695520
                ]
              ]
            },
            {
              "name": "queryExecutor",
              "columns": [
                "queriesActive",
                "queriesExecuted",
                "queriesFinished",
                "queryDurationNs"
              ],
              "values": [
                [
                  1,
                  4798,
                  4797,
                  36666623600
                ]
              ]
            },
            {
              "name": "database",
              "tags": {
                "database": "_internal"
              },
              "columns": [
                "numMeasurements",
                "numSeries"
              ],
              "values": [
                [
                  12,
                  28
                ]
              ]
            },
            {
              "name": "database",
              "tags": {
                "database": "k8s"
              },
              "columns": [
                "numMeasurements",
                "numSeries"
              ],
              "values": [
                [
                  35,
                  1937
                ]
              ]
            },
            {
              "name": "shard",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "1",
                "path": "/data/data/_internal/monitor/1",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/1"
              },
              "columns": [
                "diskBytes",
                "fieldsCreate",
                "seriesCreate",
                "writeBytes",
                "writePointsDropped",
                "writePointsErr",
                "writePointsOk",
                "writeReq",
                "writeReqErr",
                "writeReqOk"
              ],
              "values": [
                [
                  164359,
                  102,
                  28,
                  0,
                  0,
                  0,
                  29208,
                  1625,
                  0,
                  1625
                ]
              ]
            },
            {
              "name": "tsm1_engine",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "1",
                "path": "/data/data/_internal/monitor/1",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/1"
              },
              "columns": [
                "cacheCompactionDuration",
                "cacheCompactionErr",
                "cacheCompactions",
                "cacheCompactionsActive",
                "tsmFullCompactionDuration",
                "tsmFullCompactionErr",
                "tsmFullCompactions",
                "tsmFullCompactionsActive",
                "tsmLevel1CompactionDuration",
                "tsmLevel1CompactionErr",
                "tsmLevel1Compactions",
                "tsmLevel1CompactionsActive",
                "tsmLevel2CompactionDuration",
                "tsmLevel2CompactionErr",
                "tsmLevel2Compactions",
                "tsmLevel2CompactionsActive",
                "tsmLevel3CompactionDuration",
                "tsmLevel3CompactionErr",
                "tsmLevel3Compactions",
                "tsmLevel3CompactionsActive",
                "tsmOptimizeCompactionDuration",
                "tsmOptimizeCompactionErr",
                "tsmOptimizeCompactions",
                "tsmOptimizeCompactionsActive"
              ],
              "values": [
                [
                  1411406100,
                  0,
                  1,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0
                ]
              ]
            },
            {
              "name": "tsm1_cache",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "1",
                "path": "/data/data/_internal/monitor/1",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/1"
              },
              "columns": [
                "WALCompactionTimeMs",
                "cacheAgeMs",
                "cachedBytes",
                "diskBytes",
                "memBytes",
                "snapshotCount",
                "writeDropped",
                "writeErr",
                "writeOk"
              ],
              "values": [
                [
                  1368,
                  9232,
                  3972288,
                  0,
                  0,
                  0,
                  0,
                  0,
                  1625
                ]
              ]
            },
            {
              "name": "tsm1_filestore",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "1",
                "path": "/data/data/_internal/monitor/1",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/1"
              },
              "columns": [
                "diskBytes",
                "numFiles"
              ],
              "values": [
                [
                  164359,
                  1
                ]
              ]
            },
            {
              "name": "tsm1_wal",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "1",
                "path": "/data/data/_internal/monitor/1",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/1"
              },
              "columns": [
                "currentSegmentDiskBytes",
                "oldSegmentsDiskBytes",
                "writeErr",
                "writeOk"
              ],
              "values": [
                [
                  0,
                  0,
                  0,
                  1625
                ]
              ]
            },
            {
              "name": "shard",
              "tags": {
                "database": "k8s",
                "engine": "tsm1",
                "id": "2",
                "path": "/data/data/k8s/default/2",
                "retentionPolicy": "default",
                "walPath": "/data/wal/k8s/default/2"
              },
              "columns": [
                "diskBytes",
                "fieldsCreate",
                "seriesCreate",
                "writeBytes",
                "writePointsDropped",
                "writePointsErr",
                "writePointsOk",
                "writeReq",
                "writeReqErr",
                "writeReqOk"
              ],
              "values": [
                [
                  12796753,
                  1617,
                  1937,
                  0,
                  0,
                  0,
                  1380255,
                  822,
                  0,
                  822
                ]
              ]
            },
            {
              "name": "tsm1_engine",
              "tags": {
                "database": "k8s",
                "engine": "tsm1",
                "id": "2",
                "path": "/data/data/k8s/default/2",
                "retentionPolicy": "default",
                "walPath": "/data/wal/k8s/default/2"
              },
              "columns": [
                "cacheCompactionDuration",
                "cacheCompactionErr",
                "cacheCompactions",
                "cacheCompactionsActive",
                "tsmFullCompactionDuration",
                "tsmFullCompactionErr",
                "tsmFullCompactions",
                "tsmFullCompactionsActive",
                "tsmLevel1CompactionDuration",
                "tsmLevel1CompactionErr",
                "tsmLevel1Compactions",
                "tsmLevel1CompactionsActive",
                "tsmLevel2CompactionDuration",
                "tsmLevel2CompactionErr",
                "tsmLevel2Compactions",
                "tsmLevel2CompactionsActive",
                "tsmLevel3CompactionDuration",
                "tsmLevel3CompactionErr",
                "tsmLevel3Compactions",
                "tsmLevel3CompactionsActive",
                "tsmOptimizeCompactionDuration",
                "tsmOptimizeCompactionErr",
                "tsmOptimizeCompactions",
                "tsmOptimizeCompactionsActive"
              ],
              "values": [
                [
                  6420437600,
                  0,
                  5,
                  0,
                  0,
                  0,
                  0,
                  0,
                  627691500,
                  0,
                  2,
                  0,
                  283578100,
                  0,
                  1,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0
                ]
              ]
            },
            {
              "name": "tsm1_cache",
              "tags": {
                "database": "k8s",
                "engine": "tsm1",
                "id": "2",
                "path": "/data/data/k8s/default/2",
                "retentionPolicy": "default",
                "walPath": "/data/wal/k8s/default/2"
              },
              "columns": [
                "WALCompactionTimeMs",
                "cacheAgeMs",
                "cachedBytes",
                "diskBytes",
                "memBytes",
                "snapshotCount",
                "writeDropped",
                "writeErr",
                "writeOk"
              ],
              "values": [
                [
                  6410,
                  6982725,
                  18927024,
                  0,
                  3157056,
                  0,
                  0,
                  0,
                  822
                ]
              ]
            },
            {
              "name": "tsm1_filestore",
              "tags": {
                "database": "k8s",
                "engine": "tsm1",
                "id": "2",
                "path": "/data/data/k8s/default/2",
                "retentionPolicy": "default",
                "walPath": "/data/wal/k8s/default/2"
              },
              "columns": [
                "diskBytes",
                "numFiles"
              ],
              "values": [
                [
                  2704093,
                  2
                ]
              ]
            },
            {
              "name": "tsm1_wal",
              "tags": {
                "database": "k8s",
                "engine": "tsm1",
                "id": "2",
                "path": "/data/data/k8s/default/2",
                "retentionPolicy": "default",
                "walPath": "/data/wal/k8s/default/2"
              },
              "columns": [
                "currentSegmentDiskBytes",
                "oldSegmentsDiskBytes",
                "writeErr",
                "writeOk"
              ],
              "values": [
                [
                  10092660,
                  0,
                  0,
                  822
                ]
              ]
            },
            {
              "name": "shard",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "3",
                "path": "/data/data/_internal/monitor/3",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/3"
              },
              "columns": [
                "diskBytes",
                "fieldsCreate",
                "seriesCreate",
                "writeBytes",
                "writePointsDropped",
                "writePointsErr",
                "writePointsOk",
                "writeReq",
                "writeReqErr",
                "writeReqOk"
              ],
              "values": [
                [
                  341910,
                  153,
                  28,
                  0,
                  0,
                  0,
                  59887,
                  2604,
                  0,
                  2604
                ]
              ]
            },
            {
              "name": "tsm1_engine",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "3",
                "path": "/data/data/_internal/monitor/3",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/3"
              },
              "columns": [
                "cacheCompactionDuration",
                "cacheCompactionErr",
                "cacheCompactions",
                "cacheCompactionsActive",
                "tsmFullCompactionDuration",
                "tsmFullCompactionErr",
                "tsmFullCompactions",
                "tsmFullCompactionsActive",
                "tsmLevel1CompactionDuration",
                "tsmLevel1CompactionErr",
                "tsmLevel1Compactions",
                "tsmLevel1CompactionsActive",
                "tsmLevel2CompactionDuration",
                "tsmLevel2CompactionErr",
                "tsmLevel2Compactions",
                "tsmLevel2CompactionsActive",
                "tsmLevel3CompactionDuration",
                "tsmLevel3CompactionErr",
                "tsmLevel3Compactions",
                "tsmLevel3CompactionsActive",
                "tsmOptimizeCompactionDuration",
                "tsmOptimizeCompactionErr",
                "tsmOptimizeCompactions",
                "tsmOptimizeCompactionsActive"
              ],
              "values": [
                [
                  4386720800,
                  0,
                  4,
                  0,
                  1272746700,
                  0,
                  1,
                  0,
                  113799600,
                  0,
                  1,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0
                ]
              ]
            },
            {
              "name": "tsm1_cache",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "3",
                "path": "/data/data/_internal/monitor/3",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/3"
              },
              "columns": [
                "WALCompactionTimeMs",
                "cacheAgeMs",
                "cachedBytes",
                "diskBytes",
                "memBytes",
                "snapshotCount",
                "writeDropped",
                "writeErr",
                "writeOk"
              ],
              "values": [
                [
                  4351,
                  6982464,
                  8415344,
                  0,
                  0,
                  0,
                  0,
                  0,
                  2604
                ]
              ]
            },
            {
              "name": "tsm1_filestore",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "3",
                "path": "/data/data/_internal/monitor/3",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/3"
              },
              "columns": [
                "diskBytes",
                "numFiles"
              ],
              "values": [
                [
                  341910,
                  2
                ]
              ]
            },
            {
              "name": "tsm1_wal",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "3",
                "path": "/data/data/_internal/monitor/3",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/3"
              },
              "columns": [
                "currentSegmentDiskBytes",
                "oldSegmentsDiskBytes",
                "writeErr",
                "writeOk"
              ],
              "values": [
                [
                  0,
                  0,
                  0,
                  2604
                ]
              ]
            },
            {
              "name": "shard",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "4",
                "path": "/data/data/_internal/monitor/4",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/4"
              },
              "columns": [
                "diskBytes",
                "fieldsCreate",
                "seriesCreate",
                "writeBytes",
                "writePointsDropped",
                "writePointsErr",
                "writePointsOk",
                "writeReq",
                "writeReqErr",
                "writeReqOk"
              ],
              "values": [
                [
                  3695573,
                  202,
                  28,
                  0,
                  0,
                  0,
                  19539,
                  698,
                  0,
                  698
                ]
              ]
            },
            {
              "name": "tsm1_engine",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "4",
                "path": "/data/data/_internal/monitor/4",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/4"
              },
              "columns": [
                "cacheCompactionDuration",
                "cacheCompactionErr",
                "cacheCompactions",
                "cacheCompactionsActive",
                "tsmFullCompactionDuration",
                "tsmFullCompactionErr",
                "tsmFullCompactions",
                "tsmFullCompactionsActive",
                "tsmLevel1CompactionDuration",
                "tsmLevel1CompactionErr",
                "tsmLevel1Compactions",
                "tsmLevel1CompactionsActive",
                "tsmLevel2CompactionDuration",
                "tsmLevel2CompactionErr",
                "tsmLevel2Compactions",
                "tsmLevel2CompactionsActive",
                "tsmLevel3CompactionDuration",
                "tsmLevel3CompactionErr",
                "tsmLevel3Compactions",
                "tsmLevel3CompactionsActive",
                "tsmOptimizeCompactionDuration",
                "tsmOptimizeCompactionErr",
                "tsmOptimizeCompactions",
                "tsmOptimizeCompactionsActive"
              ],
              "values": [
                [
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0,
                  0
                ]
              ]
            },
            {
              "name": "tsm1_cache",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "4",
                "path": "/data/data/_internal/monitor/4",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/4"
              },
              "columns": [
                "WALCompactionTimeMs",
                "cacheAgeMs",
                "cachedBytes",
                "diskBytes",
                "memBytes",
                "snapshotCount",
                "writeDropped",
                "writeErr",
                "writeOk"
              ],
              "values": [
                [
                  0,
                  6979001,
                  0,
                  0,
                  2802384,
                  0,
                  0,
                  0,
                  698
                ]
              ]
            },
            {
              "name": "tsm1_filestore",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "4",
                "path": "/data/data/_internal/monitor/4",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/4"
              },
              "columns": [
                "diskBytes",
                "numFiles"
              ],
              "values": [
                [
                  0,
                  0
                ]
              ]
            },
            {
              "name": "tsm1_wal",
              "tags": {
                "database": "_internal",
                "engine": "tsm1",
                "id": "4",
                "path": "/data/data/_internal/monitor/4",
                "retentionPolicy": "monitor",
                "walPath": "/data/wal/_internal/monitor/4"
              },
              "columns": [
                "currentSegmentDiskBytes",
                "oldSegmentsDiskBytes",
                "writeErr",
                "writeOk"
              ],
              "values": [
                [
                  3695573,
                  0,
                  0,
                  698
                ]
              ]
            },
            {
              "name": "write",
              "columns": [
                "pointReq",
                "pointReqLocal",
                "req",
                "subWriteDrop",
                "subWriteOk",
                "writeDrop",
                "writeError",
                "writeOk",
                "writeTimeout"
              ],
              "values": [
                [
                  1488889,
                  1488889,
                  5749,
                  0,
                  5749,
                  0,
                  0,
                  5749,
                  0
                ]
              ]
            },
            {
              "name": "subscriber",
              "columns": [
                "createFailures",
                "pointsWritten",
                "writeFailures"
              ],
              "values": [
                [
                  0,
                  0,
                  0
                ]
              ]
            },
            {
              "name": "cq",
              "columns": [
                "queryFail",
                "queryOk"
              ],
              "values": [
                [
                  0,
                  0
                ]
              ]
            },
            {
              "name": "httpd",
              "tags": {
                "bind": ":8086"
              },
              "columns": [
                "authFail",
                "clientError",
                "pingReq",
                "pointsWrittenDropped",
                "pointsWrittenFail",
                "pointsWrittenOK",
                "queryReq",
                "queryReqDurationNs",
                "queryRespBytes",
                "req",
                "reqActive",
                "reqDurationNs",
                "serverError",
                "statusReq",
                "writeReq",
                "writeReqActive",
                "writeReqBytes",
                "writeReqDurationNs"
              ],
              "values": [
                [
                  0,
                  79,
                  7,
                  0,
                  0,
                  1380255,
                  4877,
                  63098601000,
                  71711123,
                  5711,
                  1,
                  86574564100,
                  0,
                  0,
                  822,
                  0,
                  415250985,
                  17761404800
                ]
              ]
            }
          ],
          "Messages": null
        }
      ]
    }
  }
]
`)
