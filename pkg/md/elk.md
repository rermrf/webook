# 什么是 ELK？
ELK 是一个强大的开源日志管理和分析平台，由三个核心组件组成：
- **Elasticsearch：分布式搜索引擎，专注于实时数据搜索和分析**。它提供了快速、灵活、可扩展的搜索和分析引擎。
- **Logstash：日志数据的收集、处理和传输工具**。Logstash 能够从多种来源收集数据，进行过滤、转换，并将数据发送到各种目的地。
- **Kibana：数据可视化工具，用于实时分析和交互式搜索**。Kibana 能够帮助用户创建动态仪表盘，展示 Elasticsearch 中存储的数据。

# ELK 的应用领域
- **日志分析**：追踪应用程序和系统的日志，帮助诊断问题和优化性能。
- **实时监控**：通过对实时数据的分析，及时发现和解决问题。
- **安全分析**：监测潜在的安全威胁和异常行为。
- **业务智能**：利用数据可视化分析，帮助业务决策。

这一切可以总结为四个字：**文本分析**。
在实践中，对于一个程序员来说，最重要的两个功能就是日志分析和实时监控。

# Logstash 
用途：Logstash 是 ELK 中的数据处理引擎，负责日志数据的收集、过滤、转换和传输。

核心功能：
- **输入（Input）**：从不同来源（如文件、数据库、消息队列）接收数据。
- **过滤（Filter）**：对接收到的数据进行解析、结构化和过滤。
- **输出（Output）**：将处理后的数据发送到指定的目的地（如 Elasticsearch）。


## Logstash 的配置文件
使用 Logstash 的核心就是要提供一份配置文件。

这个配置文件需要告诉 Logstash 从哪里读取输入，怎么处理输入，以及最终的输出是什么。


## Logstash 的过滤功能：Grok 插件
Logstash 本身支持非常多的过滤功能的插件，主要包括：
- **Grok 插件**
- **Date 插件**
- **Mutate 插件**
- **Kv 插件**


### Date 插件
**Date 插件用于将字符串转换为日期格式**，通常与 Grok 插件结合使用，将日志的时间戳字段进行解析和标准化。
```Logstash
filter {
    date {
        match => ["timestamp", "IS08601"]
    }
}
```
如上例：
- timestamp：是之前 Grok 插件解析后得到的时间戳字段。
- ISO8601：指定时间戳得格式，以便正确解析并存储。


### Mutate 插件
**Mutate 插件提供一系列数据变换操作，如重命名字段、拼接字段等.**

```conf
filter {
    mutate {
        add_field => {"new_field" => "Hello, World!"}
        remove_field => ["unwanted_field"]
        rename => {"old_field" => "new_field"}
    }
}
```

### Kv 插件
**Kv 插件用于从未结构化的文本数据中提取键值对，它可以将文本数据中的键值对解析并存储为字段。**

一般在结构化的日志里面（比如说我们的logger.Field这种）不太用得上。

```conf
filter {
  kv {
    source => "message"
    field_split => ","
  }
}
```
如上例：
- source：指定包含未结构化键值对的字段，这里使用 message 字段。
- field_split：指定键值对的分割符，这里使用逗号。


# Kibana 
**Kibana 是 ELK 中的数据可视化工具，通过直观的用户界面帮助用户查询、分析和可视化 Elasticsearch 中的数据。**

核心功能：
- **仪表盘（Dashboard）**：创建交互式仪表板，集成多个图表和可视化组件。
- **搜索和过滤**：在数据集中执行高级搜索和过滤，以定位感兴趣的数据。
- **图标和可视化**：使用多种图表类型，如柱状图、折线图、地图等，呈现数据。


## Kibana 的使用场景
主要应用场景有:
- **日志分析**：通过 Kibana 对 Elasticsearch 中的日志数据进行搜索、过滤和可视化，实时监控系统运行状况。
- **性能监控**：利用 Kibana 的仪表板功能展示系统性能指标，帮助发现和解决性能问题。
- **安全分析**：使用 Kibana 进行安全事件的可视化分析，提高对潜在威胁的识别和响应能力。

简单来说，Kibana 是一个侧重数据展示的框架，主打的就是四个字，**花里胡哨**.

## Kibana VS Grafana
Grafana 和 Kibana 都是流行的数据可视化框架，他们都提供了强大的可视化工具。但是它们在设计目标、适用场景和特性方面存在一些区别：
Kibana：
- 设计目标：主要设计用于与 Elasticsearch 协同工作，构建实时搜索和分析的仪表板。Kibana 深度集成了 ELK 堆栈，专注于日志和指标数据的可视化。
- 告警和通知：提供告警功能，可以设置基于查询结果的告警规则。然而，Kibana 的告警功能相对较新，并且相对较简单。

Grafana:
- 设计目标：最初设计为通用的开源仪表板和可视化平台，支持各种数据源。
- 告警和通知：拥有更强大和成熟的告警系统，支持多种通知渠道，例如电子邮件、Slack、Webhooks等。Grafana 的告警功能更加灵活，允许定义复杂的触发条件。



# 部署 ELK
```yaml
  elasticsearch:
    image: elasticsearch:8.15.5
    container_name: elasticsearch
    environment:
      # 单节点形态
      - "discovery.type=single-node"
      # 禁用 xpack 功能
      - "xpack.security.enabled=false"
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
  logstash:
    image: elastic/logstash:8.15.5
    container_name: logstash
    volumes:
      - ./config/logstash:/usr/share/logstash/pipeline
    #      - ./logstash-logs:/usr/share/logstash/logs
    #      - ./app.log:/usr/share/logstash/app.log
    environment:
      - "xpack.monitoring.elasticsearch.hosts=http://elasticsearch:9200"
    ports:
      - "5044:5044"
  kibana:
    image: docker.elastic.co/kibana/kibana:8.15.5
    container_name: kibana
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
      - i18n.locale=zh-CN
    ports:
      - "5601:5601"
```

## 日志初始化
使用 lumberjack 库管理日志切片

```go
package ioc

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"webook/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	lumberLogger := &lumberjack.Logger{
		Filename:   "E:\\app\\misc\\新建文件夹\\comment.log", // 指定日志文件路径
		MaxSize:    50,                                  // 每个日志文件的最大大小
		MaxBackups: 3,                                   // 保留旧日志的最大个数
		MaxAge:     28,                                  // 保留久日志文件的最大天数
		//Compress:   true,                                // 是否压缩旧的日志文件，测试环境下不用开
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(lumberLogger),
		zapcore.DebugLevel, // 设置日志级别
	)
	l := zap.New(core, zap.AddCaller())

	return logger.NewZapLogger(l)
}

```


### 总结：为什么要使用 Filebeat？
实践中结合 Filebeat 是很常见的做法，主要是为了简化日志收集和传输的过程，提高整个日志处理流程的效率。优点有：
- **轻量级和高效。**
- **实时性**：Filebeat 能够实时检测并收集日志文件的变化，将新的日志数据迅速传输到指定的目的地，以实现实时分析和监控。
- **模块化配置**：Filebeat 提供了模块化的配置，支持多种数据源和数据格式，包括系统日志、NGINX 日志、Apache 日志等。
- **支持多种输出**：Filebeat 可以将收集到的日志数据发送到多个不同的目的地，如 Logstash、Elasticsearch、Kafka 等，以适应不同的日志处理需求。
- **结合 Logstash**：虽然 Filebeat 可以直接将数据发送到 Elasticsearch，但通常它会与 Logstash 一起使用，以实现更复杂的数据处理、过滤和转换。这种结合使用使得整个 ELK 堆栈更加灵活。
- **自动发现和标准化**：Filebeat 支持自动发现新的日志文件，并能够标准化日志数据的格式，使其更易于搜索和分析。
- **适用于容器环境**：Filebeat 可以轻松集成到容器环境中，实现对容器内应用的日志收集，支持各种容器化平台。