// Package node 负责启动和管理以太坊节点生命周期，主要作用是：
//•	启动以太坊节点，初始化各类资源（如 RPC、p2p 网络、数据库等）。
//•	管理节点的状态：如启动、关闭、运行等。
//•	提供服务接口，允许其他模块注册（例如注册 RPC 服务）。

/*
Package node sets up multi-protocol Ethereum nodes.
// 包 node 用于设置多协议以太坊节点

In the model exposed by this package, a node is a collection of services which use shared
resources to provide RPC APIs. Services can also offer devp2p protocols, which are wired
up to the devp2p network when the node instance is started.
// 在这个包暴露的模型中，一个节点是由多个服务组成的集合，这些服务使用共享资源来提供 RPC API。服务还可以提供 devp2p 协议，当节点实例启动时，这些协议会与 devp2p 网络连接。

# Node Lifecycle
# 节点生命周期

The Node object has a lifecycle consisting of three basic states, INITIALIZING, RUNNING
and CLOSED.

	 Node 对象的生命周期包含三个基本状态：INITIALIZING、RUNNING 和 CLOSED。

		●───────┐
		     New()
		        │
		        ▼
		  INITIALIZING ────Start()─┐
		        │                  │
		        │                  ▼
		    Close()             RUNNING
		        │                  │
		        ▼                  │
		     CLOSED ◀──────Close()─┘

Creating a Node allocates basic resources such as the data directory and returns the node
in its INITIALIZING state. Lifecycle objects, RPC APIs and peer-to-peer networking
protocols can be registered in this state. Basic operations such as opening a key-value
database are permitted while initializing.
// 创建一个 Node 会分配基本资源，如数据目录，并将节点置于 INITIALIZING 状态。生命周期对象、RPC API 和点对点（p2p）网络协议可以在此状态下注册。初始化时可以执行的基本操作，如打开一个键值数据库。

Once everything is registered, the node can be started, which moves it into the RUNNING
state. Starting the node starts all registered Lifecycle objects and enables RPC and
peer-to-peer networking. Note that no additional Lifecycles, APIs or p2p protocols can be
registered while the node is running.
// 一旦所有内容都注册完毕，节点可以启动，进入 RUNNING 状态。启动节点会启动所有已注册的生命周期对象，并启用 RPC 和 p2p 网络。注意，节点运行时不能再注册新的生命周期、API 或 p2p 协议。

Closing the node releases all held resources. The actions performed by Close depend on the
state it was in. When closing a node in INITIALIZING state, resources related to the data
directory are released. If the node was RUNNING, closing it also stops all Lifecycle
objects and shuts down RPC and peer-to-peer networking.
// 关闭节点会释放所有已占用的资源。关闭操作根据节点当前状态有所不同。如果节点处于 INITIALIZING 状态，关闭时会释放与数据目录相关的资源。如果节点处于 RUNNING 状态，关闭时会停止所有生命周期对象，并关闭 RPC 和 p2p 网络。

You must always call Close on Node, even if the node was not started.
// 你必须始终调用 Close 来关闭 Node，即使该节点未启动。

# Resources Managed By Node
// ## Node 管理的资源

All file-system resources used by a node instance are located in a directory called the
data directory. The location of each resource can be overridden through additional node
configuration. The data directory is optional. If it is not set and the location of a
resource is otherwise unspecified, package node will create the resource in memory.
// 节点实例使用的所有文件系统资源都位于一个名为数据目录的目录中。每个资源的位置可以通过额外的节点配置进行覆盖。数据目录是可选的。如果未设置数据目录，且其他资源的位置未指定，则 package node 会在内存中创建资源。

To access to the devp2p network, Node configures and starts p2p.Server. Each host on the
devp2p network has a unique identifier, the node key. The Node instance persists this key
across restarts. Node also loads static and trusted node lists and ensures that knowledge
about other hosts is persisted.
// 为了访问 devp2p 网络，Node 配置并启动了 p2p.Server。devp2p 网络中的每个主机都有一个唯一的标识符，称为节点密钥。Node 实例会在重启时保留该密钥。Node 还会加载静态和受信任的节点列表，并确保有关其他主机的知识得以持久化。

JSON-RPC servers which run HTTP, WebSocket or IPC can be started on a Node. RPC modules
offered by registered services will be offered on those endpoints. Users can restrict any
endpoint to a subset of RPC modules. Node itself offers the "debug", "admin" and "web3"
modules.
// 可以在 Node 上启动运行 HTTP、WebSocket 或 IPC 的 JSON-RPC 服务器。已注册服务提供的 RPC 模块将在这些端点上提供。用户可以限制任何端点仅提供某些 RPC 模块。Node 本身提供 "debug"、"admin" 和 "web3" 模块。

Service implementations can open LevelDB databases through the service context. Package
node chooses the file system location of each database. If the node is configured to run
without a data directory, databases are opened in memory instead.
// 服务实现可以通过服务上下文打开 LevelDB 数据库。package node 选择每个数据库的文件系统位置。如果节点配置为在没有数据目录的情况下运行，则数据库会被打开在内存中。

Node also creates the shared store of encrypted Ethereum account keys. Services can access
the account manager through the service context.
// Node 还会创建加密以太坊账户密钥的共享存储。服务可以通过服务上下文访问账户管理器。

# Sharing Data Directory Among Instances
// ## 多实例共享数据目录

Multiple node instances can share a single data directory if they have distinct instance
names (set through the Name config option). Sharing behaviour depends on the type of
resource.
// 多个节点实例可以共享同一个数据目录，前提是它们有不同的实例名称（通过 Name 配置选项设置）。共享行为取决于资源的类型。

devp2p-related resources (node key, static/trusted node lists, known hosts database) are
stored in a directory with the same name as the instance. Thus, multiple node instances
using the same data directory will store this information in different subdirectories of
the data directory.
// 与 devp2p 相关的资源（节点密钥、静态/受信任的节点列表、已知主机数据库）存储在与实例同名的子目录中。因此，多个节点实例使用相同的数据目录时，这些信息会存储在数据目录的不同子目录中。

LevelDB databases are also stored within the instance subdirectory. If multiple node
instances use the same data directory, opening the databases with identical names will
create one database for each instance.
// LevelDB 数据库也存储在实例子目录中。如果多个节点实例使用相同的数据目录，打开具有相同名称的数据库会为每个实例创建一个数据库。

The account key store is shared among all node instances using the same data directory
unless its location is changed through the KeyStoreDir configuration option.
// 账户密钥存储是所有使用相同数据目录的节点实例共享的，除非通过 KeyStoreDir 配置选项改变其位置。

# Data Directory Sharing Example
// ## 数据目录共享示例

In this example, two node instances named A and B are started with the same data
directory. Node instance A opens the database "db", node instance B opens the databases
"db" and "db-2". The following files will be created in the data directory:
// 在这个示例中，两个节点实例 A 和 B 使用相同的数据目录启动。节点实例 A 打开数据库 "db"，节点实例 B 打开 "db" 和 "db-2" 数据库。将在数据目录中创建以下文件：

	data-directory/
		A/
			nodekey            -- devp2p node key of instance A
			nodes/             -- devp2p discovery knowledge database of instance A
			db/                -- LevelDB content for "db"
		A.ipc                  -- JSON-RPC UNIX domain socket endpoint of instance A
		B/
			nodekey            -- devp2p node key of node B
			nodes/             -- devp2p discovery knowledge database of instance B
			static-nodes.json  -- devp2p static node list of instance B
			db/                -- LevelDB content for "db"
			db-2/              -- LevelDB content for "db-2"
		B.ipc                  -- JSON-RPC UNIX domain socket endpoint of instance B
		keystore/              -- account key store, used by both instances
*/
package node
