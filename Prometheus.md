# K6 K9-AMQP extension prometheus remote write

## Summary

K6 allows to write K6 metrics using (prometheus-remote-write)[https://grafana.com/docs/k6/latest/results-output/real-time/prometheus-remote-write/#prometheus-remote-write] experimental module.


## Run Prometheus and Grafana

Use your Prometheus and Grafana or run locally.

### Create prometheus.yaml (empty file)


```yaml

```

### Create empty `data` directory


```sh
mkdir data

```

### Create docker-compose.yaml

```yaml
networks:
  prometheus:
    driver: bridge
  rabbitmq:
    driver: bridge

services:
  grafana:
    container_name: grafana
    image: grafana/grafana:13.1.0
    hostname: grafana
    ports:
      - "3000:3000"
    networks:
      - prometheus
  prometheus:
    container_name: prometheus
    image: prom/prometheus:v3.13.1
    hostname: prometheus
    ports:
      - "9090:9090"
    networks:
      - prometheus
    volumes:
      - type: bind
        source: ./prometheus.yaml
        target: /prometheus/prometheus.yaml
      - type: bind
        source: ./data/prometheus
        target: /test
    command: "--config.file=prometheus.yaml --storage.tsdb.retention.time=60d --web.enable-remote-write-receiver"
  rabbitmq:
    container_name: rabbitmq
    image: rabbitmq:4.3.2-management-alpine
    hostname: rabbitmq
    ports:
      - "5672:5672"
      - "15672:15672"
    networks:
      - rabbitmq
```

### Run Grafana and Prometheus


```sh
sudo docker compose up
```

```sh
[+] up 5/5
 ✔ Network k6-test-with-grafana_rabbitmq   Created                                                                                                                                                0.1s
 ✔ Network k6-test-with-grafana_prometheus Created                                                                                                                                                0.0s
 ✔ Container rabbitmq                      Created                                                                                                                                                0.1s
 ✔ Container grafana                       Created                                                                                                                                                0.0s
 ✔ Container prometheus                    Created                                                                                                                                                0.1s
Attaching to grafana, prometheus, rabbitmq
prometheus  | time=2026-10-24T12:32:18.741Z level=INFO source=main.go:1770 msg="updated GOGC" old=100 new=75
prometheus  | time=2026-10-24T12:32:18.741Z level=INFO source=main.go:824 msg="Leaving GOMAXPROCS=16: CPU quota undefined" component=automaxprocs
prometheus  | time=2026-10-24T12:32:18.741Z level=INFO source=memlimit.go:198 msg="GOMEMLIMIT is updated" component=automemlimit package=github.com/KimMachineGun/automemlimit/memlimit GOMEMLIMIT=29119592448 previous=9223372036854775807
prometheus  | time=2026-10-24T12:32:18.741Z level=INFO source=main.go:933 msg="Starting Prometheus Server" mode=server version="(version=3.15.1, branch=HEAD, revision=bae3eced0a5369e050e83ea73bc381d7c4b1a719)"
prometheus  | time=2026-10-24T12:32:18.741Z level=INFO source=main.go:938 msg="operational information" build_context="(go=go1.27.3, platform=linux/amd64, user=root@03bc6b0f3ef7, date=20261021-13:31:17, tags=netgo,builtinassets)" host_details="(Linux 7.3.0-9-generic #9-Ubuntu SMP PREEMPT_DYNAMIC Mon Oct  5 10:38:02 UTC 2026 x86_64 prometheus (none))" fd_limits="(soft=524287, hard=524288)" vm_limits="(soft=unlimited, hard=unlimited)"
grafana     | logger=settings t=2026-10-24T12:32:18.758704342Z level=info msg="Starting Grafana" version=13.2.4 commit=0e79aef9d91abf3bdadae9af15763029072fb45b branch=release-13.2.4#patched compiled=2026-10-12T12:57:33Z
grafana     | logger=settings t=2026-10-24T12:32:18.759182718Z level=info msg="Unified migration configs enforced" storage_type=unified target=[all]
grafana     | logger=settings t=2026-10-24T12:32:18.759194781Z level=info msg="Enforcing mode 5 for resource in unified storage" resource=folders.folder.grafana.app
grafana     | logger=settings t=2026-10-24T12:32:18.759198177Z level=info msg="Enforcing mode 5 for resource in unified storage" resource=shorturls.shorturl.grafana.app
grafana     | logger=settings t=2026-10-24T12:32:18.759200081Z level=info msg="Enforcing mode 5 for resource in unified storage" resource=playlists.playlist.grafana.app
grafana     | logger=settings t=2026-10-24T12:32:18.759203577Z level=info msg="Enforcing mode 5 for resource in unified storage" resource=dashboards.dashboard.grafana.app
grafana     | logger=settings t=2026-10-24T12:32:18.759290831Z level=info msg="Config loaded from" file=/usr/share/grafana/conf/defaults.ini
grafana     | logger=settings t=2026-10-24T12:32:18.759293175Z level=info msg="Config loaded from" file=/etc/grafana/grafana.ini
grafana     | logger=settings t=2026-10-24T12:32:18.759294929Z level=info msg="Config overridden from command line" arg="default.paths.data=/var/lib/grafana"
grafana     | logger=settings t=2026-10-24T12:32:18.759296482Z level=info msg="Config overridden from command line" arg="default.paths.logs=/var/log/grafana"
grafana     | logger=settings t=2026-10-24T12:32:18.759298044Z level=info msg="Config overridden from command line" arg="default.paths.plugins=/var/lib/grafana/plugins"
grafana     | logger=settings t=2026-10-24T12:32:18.759299507Z level=info msg="Config overridden from command line" arg="default.paths.provisioning=/etc/grafana/provisioning"
grafana     | logger=settings t=2026-10-24T12:32:18.759302763Z level=info msg="Config overridden from command line" arg="default.log.mode=console"
grafana     | logger=settings t=2026-10-24T12:32:18.759304547Z level=info msg="Config overridden from Environment variable" var="GF_PATHS_DATA=/var/lib/grafana"
grafana     | logger=settings t=2026-10-24T12:32:18.75930606Z level=info msg="Config overridden from Environment variable" var="GF_PATHS_LOGS=/var/log/grafana"
grafana     | logger=settings t=2026-10-24T12:32:18.759307693Z level=info msg="Config overridden from Environment variable" var="GF_PATHS_PLUGINS=/var/lib/grafana/plugins"
grafana     | logger=settings t=2026-10-24T12:32:18.759309225Z level=info msg="Config overridden from Environment variable" var="GF_PATHS_PROVISIONING=/etc/grafana/provisioning"
grafana     | logger=settings t=2026-10-24T12:32:18.759310638Z level=info msg="Config overridden from Environment variable" var="GF_PATHS_HOME=/usr/share/grafana"
grafana     | logger=settings t=2026-10-24T12:32:18.759312622Z level=info msg="Config overridden from Environment variable" var="GF_PATHS_CONFIG=/etc/grafana/grafana.ini"
grafana     | logger=settings t=2026-10-24T12:32:18.759314285Z level=info msg=Target target=[all]
grafana     | logger=settings t=2026-10-24T12:32:18.759316529Z level=info msg="Path Home" path=/usr/share/grafana
grafana     | logger=settings t=2026-10-24T12:32:18.759318212Z level=info msg="Path Data" path=/var/lib/grafana
grafana     | logger=settings t=2026-10-24T12:32:18.759321328Z level=info msg="Path Logs" path=/var/log/grafana
grafana     | logger=settings t=2026-10-24T12:32:18.759325756Z level=info msg="Path Plugins" path="[/var/lib/grafana/plugins /usr/share/grafana/data/plugins-bundled]"
grafana     | logger=settings t=2026-10-24T12:32:18.759328101Z level=info msg="Path Provisioning" path=/etc/grafana/provisioning
grafana     | logger=settings t=2026-10-24T12:32:18.759329674Z level=info msg="App mode production"
grafana     | logger=featuremgmt t=2026-10-24T12:32:18.759897889Z level=info msg=FeatureToggles alertRuleRestore=true alertingBulkActionsInUI=true alertingImportYAMLUI=true alertingListViewV2=true alertingMigrationUI=true alertingNavigationV2=true alertingNotificationsStepMode=true alertingQueryAndExpressionsStepMode=true alertingRulePermanentlyDelete=true alertingRuleRecoverDeleted=true alertingRuleVersionHistoryRestore=true alertingUIOptimizeReducer=true alertingUIUseBackendFilters=true alertingUIUseFullyCompatBackendFilters=true alertingUseNewSimplifiedRoutingHashAlgorithm=true annotationPermissionUpdate=true auditLoggingAppPlatform=true awsAsyncQueryCaching=true awsDatasourcesTempCredentials=true azureMonitorEnableUserAuth=true azureMonitorPrometheusExemplars=true azureResourcePickerUpdates=true cloudWatchCrossAccountQuerying=true cloudWatchNewLabelParsing=true cloudWatchRoundUpEndTime=true dashboardNewLayouts=true dashboardUnifiedDrilldownControls=true disableScriptedDashboards=true enableSCIM=true feedbackButton=true grafana.dashboardSettingsRedesign=true grafana.newPreferencesPage=true grafana.queryVarEditorRedesign=true grafana.scenesFlickeringFix=true grafana.unifiedDataSourcePicker=true grafana.viewPanelPane=true grafanaAssistantInProfilesDrilldown=true improvedExternalSessionHandling=true improvedExternalSessionHandlingSAML=true lokiLabelNamesQueryApi=true lokiQuerySplitting=true newClickhouseConfigPageDesign=true newSavedQueriesExperience=true onlyStoreActionSets=true preferences.rerouteLegacyAPIs=true profilesExemplars=true prometheusAzureOverrideAudience=true prometheusTypeMigration=true provisioning.gitConventions=true provisioning.readmes=true provisioning.userAttribution=true provisioningFolderMetadata=true pyroscopeUTF8LabelNames=true queryLibrary=true react19=true rememberUserOrgForSso=true renderAuthJWT=true restrictedPluginApis=true savedQueriesRBAC=true sqlExpressions=true useKubernetesShortURLsAPI=true useSessionStorageForRedirection=true
grafana     | logger=sqlstore t=2026-10-24T12:32:18.759940689Z level=info msg="Connecting to DB" dbtype=sqlite3
grafana     | logger=sqlstore t=2026-10-24T12:32:18.759944266Z level=info msg="Using SQLite driver" driver=modernc.org/sqlite
grafana     | logger=sqlstore t=2026-10-24T12:32:18.75995203Z level=info msg="Creating SQLite database file" path=/var/lib/grafana/grafana.db
grafana     | logger=migrator t=2026-10-24T12:32:18.760680897Z level=info msg="Starting DB migrations"
grafana     | logger=migrator t=2026-10-24T12:32:18.761261395Z level=info msg="Executing migration" id="create migration_log table"
...
```

### Check Prometheus, Grafana and RabbitMQ are up

[Local Prometheus](http://localhost:9090)
[Local Grafana](http://localhost:9090)
[Local RabbitMQ](http://localhost:15672/)

### Add Prometheus Data Source to grafana

[Add Data Source](http://localhost:3000/connections/datasources/new)

Use `http://prometheus:9090` URL as Prometheus Server URL and `Save & test`

### Check Grafana Metrics

[Grafana Metrics](http://localhost:3000/a/grafana-metricsdrilldown-app/drilldown)

### Run K6 Test

Run K6 test using `K6_PROMETHEUS_RW_SERVER_URL` environment variable.


```sh
K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write k6 run -o experimental-prometheus-rw ./examples/produce-consume.js
```

```sh
INFO[0000] 2026/07/12 12:06:45 INFO init amqp client with pool {ChannelsPerConn:1 ChannelsCacheSize:20}

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /  
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/


     execution: local
        script: ./examples/produce-consume.js
        output: Prometheus remote write (http://localhost:9090/api/v1/write)

     scenarios: (100.00%) 2 scenarios, 50 max VUs, 1m40s max duration (incl. graceful stop):
              * publish: 333.33 iterations/s for 1m0s (maxVUs: 20-40, exec: produce, gracefulStop: 30s)
              * consume: 1.67 iterations/s for 1m0s (maxVUs: 5-10, exec: consume, startTime: 10s, gracefulStop: 30s)

INFO[0000] 2026/07/12 12:06:45 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 12:06:45 INFO exchange created name=test.ex
INFO[0000] 2026/07/12 12:06:45 INFO queue created name=test.q
INFO[0000] 2026/07/12 12:06:45 INFO queue binded name=test.q key=test
INFO[0010] 2026/07/12 12:06:55 INFO no available channel in pool, creating new one
INFO[0070] 2026/07/12 12:07:55 INFO Teardown AMQP Client


  █ TOTAL RESULTS 

    CUSTOM
    amqp_pub_failed........: 0     0/s
    amqp_pub_sent..........: 20001 285.559672/s
    amqp_sub_failed........: 0     0/s
    amqp_sub_no_delivery...: 96    1.370618/s
    amqp_sub_received......: 4     0.057109/s

    EXECUTION
    iteration_duration.....: avg=137µs min=17.81µs med=85.69µs max=8.96ms p(90)=291.56µs p(95)=328.98µs
    iterations.............: 20101 286.987399/s
    vus....................: 0     min=0        max=1
    vus_max................: 25    min=25       max=25

    NETWORK
    data_received..........: 0 B   0 B/s
    data_sent..............: 0 B   0 B/s




running (1m10.0s), 00/25 VUs, 20101 complete and 0 interrupted iterations
publish ✓ [======================================] 00/20 VUs  1m0s  333.33 iters/s
consume ✓ [======================================] 00/05 VUs  1m0s  1.67 iters/s
```

### Check Promethues Metrics in Grafana

![Grfana](./Grafana-metrics.png)

![Prometheus - publish](./Prometheus-metrics-pub.png)

![Prometheus - consume](./Prometheus-metrics-cons.png)