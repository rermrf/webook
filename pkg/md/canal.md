# Canal

## Canal 是什么
- **Canal 是一款开源的数据库实时变更监控和数据同步工具**，支持多种数据库系统，包括 Mysql、MariaDB、阿里云 RDS 等。
- **它允许实时捕获数据库的变更，提供高性能的数据同步服务**，被广泛应用于数据仓库同步、实时分析等场景。

或者说，Canal 是一个典型的 CDC（changge data capture）工具。

优势：
- **实时性高**：Canal 提供实时的数据变更监控和同步，使得系统能够更迅速地响应数据库的变更。
- **灵活性高**：Canal 的配置和使用相对简单，可以轻松应对不同场景的数据同步需求。
- **开源社区支持**：由于是开源工具，Canal 受到了活跃的开源社区支持，可以获取到及时的更新和问题解决方案。


## Canal 的作用
- **实时监控数据库变更**：Canal 允许你实时监控数据库的变更，包括插入、更新和删除操作。这种实时性对于需要及时了解数据库变更的应用场景非常重要。
- **数据同步**：Canal 提供了高效的数据同步机制，可以将一个数据库的变更同步到另一个数据库，保持数据一致性。这对于分布式系统中的数据同步是至关重要的。
- **支持实时分析**：Canal 的实时性使得数据能够被及时传输到数据仓库或分析平台，支持实时数据分析和报告生成。这对于需要快速响应数据变化的业务非常有用。
- **解耦数据库系统**：使用 Canal 可以帮助解耦不同数据系统之间的依赖关系。这意味着你可以在系统中引入新的数据库或更改数据库结构，而不必担心影响到其他部分的运作。


## Canal 的基本组成
- **Canal Server**：基本组件之一，负责连接到数据库，并实时监控数据库的变更。它捕获变更日志并将其发送给连接的客户端。
- **Canal Client**：Canal Client 是与 Canal Server 进行通信的组件，用于接收并处理从 Canal Server 发送过来的数据库变更。应用程序可以通过 Canal Client 获取实时的数据库变更信息。
- **Binlog**：Canal 使用数据库的二进制日志进行实时监控和捕获变更。
- **数据格式转换器**：Canal 支持多种数据格式，例如 JSON、Avro 等。数据格式转换器负责将从数据库捕获的变更日志转换为用户指定的数据格式。
- **Canal 配置文件**：Canal 的配置文件包含了与数据库连接、监控规则、数据格式等相关的配置信息。通过配置文件，用户可以灵活地定制 Canal 的行为。
- **ZooKeeper（可选）**：在分布式场景中，Canal 可以使用 Zookeeper 来进行服务的协调和管理。Zookeeper 提供了高可用和容错性的支持。

<img src="./img/canal组成.png">


## 什么是 Binlog？
Binlog 是数据库中的二进制日志，**记录数据库中的每个变更操作**。它包含了对数据库进行插入、更新和删除的详细信息。

它是 Canal 实时捕获变更的重要基础。

<img src="./img/binlog主从同步.png">

如上图展示了 Binlog 用于主从同步。基本上就是四个步骤：
- 从库连上主库。
- 从库发起数据同步。
- 主库开启一个线程，Binlog发送到从节点。
- 从节点收到 Binlog，先写到 Relay log，而后逐步执行 Relay log 中的数据变更。

### Relay log
- 记录同步过来的每一个操作，如果从库崩了也能恢复
- 平衡主库和从库之间的速率


### Binlog 的三种模式
- **Row-based logging(基于行的日志记录)**：记录每行数据的变更，适用于那些以行为单位进行数据变更的场景。这种模式提供了最详细的变更信息，但可能会产生较大的日志量。
- **Statement-based Logging(基于语句的日志记录)**：记录 SQL 语句的执行，以表达数据变更的操作。这种模式适用于那些以 SQL 语句为单位进行数据变更的场景。它生成的日志量相对较小，但可能无法捕获一些复杂的数据变更情况。
- **Mixed Logging(混合模式日志记录)**：结合了行级和语句级两种日志记录模式的优势。在这种模式下，数据库根据具体的数据变更情况使用行级或语句级记录。这种模式平衡了详细信息和日志量之间的权衡。

Canal 支持解析和处理这三种不同的 Binlog 模式，因此用户可以根据实际情况选择适合其应用的模式。
个人认为，基于行的日志记录用起来方便。


## Canal 的配置
Canala 的配置比较复杂，需要配置的东西也比较多。
它的配置大体上分成三个部分：
- **Canal Server 本体的配置**。也就是 Canal Server 自己运作需要的配置。
- **连上不同数据库的配置**。每次你要连一个不同的数据库，你就需要提供一份配置。这一份配置的关键就是提供连上数据库必要的连接信息、用户信息。
- **转发配置**。也就是 Canal 收到了 Binlog 之后，要把这个数据转发到哪里。

```toml
canal.admin.port = 11110
canal.admin.user = admin
canal.
```


# canal 使用案例


## 案例一：借助 Canal 来更新缓存
正常我们在使用缓存的时候，都会面临更新缓存和更新 DB 的并发问题。

那么这里我们可以考虑借助 Canal 来更新缓存。

但是在更新缓存的时候，有两种做法：
- **直接用 Canal 里面的数据来更新**。这种做法的关键点，就是要配置好 Canal 的消息，确保同一条数据的 Binlog 一定会在同一个 Kafka 分区上。
- **Canal 只被用作一个信号器，数据从数据库里面加载，再回写**。这种机制性能比较差，对数据库压力比较大。


### 消息顺序问题
走第一条路，那么就需要保证消息顺序，在 Canal 的文档里面提供了控制 topic 和分区的配置。

```properties
# mq config
canal.mq.topic=webook_binlog
# dynamic topic route by schema or table regex
#canal.mq.dynamicTopic=*\\..*
#canal.mq.partition=0
# hash partition config
#canal.mq.enableDynamicQueuePartition=false
canal.mq.partitionsNum=3
#canal.mq.dynamicTopicPartitionNum=test.*:4,mycanal:6
# 按照 id 来哈希
canal.mq.partitionHash=.*\\..*:id
```

### 表和 topic 的关系
在大厂里面，或者说在高并发大数据的应用里面，如果所有的 MYSQL 的表都在同一个 topic，那么就容易出现：
- **消息积压。**
- **Kafka 集群性能瓶颈。**

所以在大规模应用里面，你要考虑：
- 使用不同的集群。
- 使用不同的 topic。例如说比较常见的就是不同的表使用不同的 topic。


### 代码实现：
这里有一个比较关键的点：**利用 Binlog 来更新缓存，是缓存策略，而不是业务逻辑家**。

所以一般不会通过 Service 来更新，而是绕开 Service，操作 Repository。

在部分情况下，比如说 Repository 本身也没什么逻辑，那么可以直接操作 Cache。

而且，这是具体缓存策略，所以并不适合在 Repository 上定义相关的接口。


```go
type MysqlBinlogConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	//耦合到实现，而不是耦合到接口，除非你把操作缓存的方法，也定义到 repository 接口上
	repo *repository.CachedArticleRepository
}

func (m *MysqlBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("public_article_cache", m.client)
	if err != nil {
		return err
	}
	go func() {
		// 这里逼不得已和 DAO 耦合在一起
		err := cg.Consume(context.Background(), []string{"webook_binlog"}, saramax.NewHandler[canalx.Message[dao.PublishedArticle]](m.l, m.Consume))
		if err != nil {
			m.l.Error("退出了消费循环", logger.Error(err))
		}
	}()
	return err
}

func (m *MysqlBinlogConsumer) Consume(msg *sarama.ConsumerMessage, val canalx.Message[dao.PublishedArticle]) error {
	// 别的表的 binlog，不需要关心
	// 可以考虑不同的表用不同的 topic ，那么这里就不需要判定了
	if val.Table != "published_articles" {
		return nil
	}

	// 更新缓存
	// 增删改的消息，实际上在 publish article 里面是没有删的消息
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, data := range val.Data {
		var err error
		switch data.Status {
		case domain.ArticleStatusPublished.ToUint8():
			// 发表要写入
			err = m.repo.Cache().SetPub(ctx, m.repo.ToDomain(dao.Article(data)))
		case domain.ArticleStatusPrivate.ToUint8():
			err = m.repo.Cache().DelPub(ctx, data.Id)
		}
		if err != nil {
			// 记录日志就行
			m.l.Error("使用canal通知缓存修改失败", logger.Error(err))
		}
	}
	return nil
}

```


## 案例二：借助 Canal 完成数据校验
在数据迁移那个部分，我们提到可以使用 Canal 来完成增量的数据校验与修复。

业务方只需要考虑创建出来消费者，调用这个方法就行。