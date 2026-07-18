#分布式系统 

参考资料：MIT Distributed System 6.5840 Lab 2: Key/Value Server; Spring 2026

这周完成了MIT Distributed System的Lab 2，趁热打铁，把自己的思考与实现过程记录下来。

# 引言

在真正设计key-value server之前，我们先看一下如何在本地的电脑做一个key-value storage（类似Python的字典）（不适用多进程与线程）。

首先，我们需要设计一个储存在内存中的哈希表：
```
map[string]string
```

然后我们可以设计类似`Put(key, value)`与`Get(key)`的函数接口来修改或者获得哈希表中的数据。

因为是单机上的单一进程与线程，所以不会存在同时有两个`Put`或者`Get`需要处理，所有的request都是依次顺序抵达哈希表，然后按照到达的顺序处理记即可。非常直观，完全依赖数据结构与算法里学过的内容即可解决:
```
map = {}

Put("A", "100") --> map = {"A": "100"}
Put("A", "999") --> map = {"A": "999"}
Get("A") --> 999
```

此时，我们引入分布式系统：假使我们的key-value store由一个服务器所维护，而`Put`与`Get`操作则由不同的客户端发起。因为服务器与客户端分属于不同的机器，他们之间不会共享内存，因此客户不可能直接修改或者访问key-value store。此时我们需要依赖于RPC通信协议，构成如下的通信方式：

```					
client A -- RPC ---> Sever <-- RPC --- Client B
						|
						|
				---------------------
				|   Key-Value Store |
				---------------------
```

在这情况架构下，我们会遇到如下问题：
* 如果客户发送`Put`与`Get`请求没有送达服务器端怎么办（网络通信并不可靠）？
* 如果服务器收到了客户的请求，但是发给客户的回执丢失怎么办？
* 如果两个客户同时都要对同一个key的value做出修改，怎么保证同一时间只有一个客户端可以成功？
* 从客户端的视角来看，发送一个`Put`请求没有收到服务器的回执，可能意味着请求没有送到，也可能意味着请求送到服务器，`Put`操作也成功，但是回执没有收到。客户端在遇到这样面临模棱两可的任务该怎么处理？

相信做完Lab 2后会对以上问题有比较深刻的理解。

# 系统概述

在Lab 2里，我们最终需要实现的系统如下图所示：

```
multiple clients -- unreliable RPC call --> Server
                                             |
                                             |
                                    --------------------
			                        |  Key Value Store |
			                        --------------------
```


我们在整个Lab里只需要实现三类操作：
* Put：对key value进行修改
* Get：根据给定的key，返回对应的value（或者错误信息）
* Lock：通过key-value数据结构实现分布式锁

# 具体实现

这门课的Lab已经贴心的将一个相对较复杂的任务切分成了四个小任务，我们在实现的时候可以依次按照提示完成四个阶段的任务。

我在复盘时，直接跳过四个阶段，转而分别从服务器端，客户端，以及锁Lock三个层面的实现来阐述，这样能更好的理解这个Lab带给我的思考。
## 服务器端的实现

### 数据结构设计

在服务器端，我们需要思考让服务器保存什么样的变量才可以实现key-value store：

首先我们自然想到为了防止race condition出现，我们需要引入mutex（这一点与多线程问题类似）。

其次与单机相似的是，我们仍旧需要一个哈希表来储存key-value pair，这是核心变量。

仅靠以上两个变量足够实现分布式系统的key value server吗？让我们来看下面这个场景：
```
# 两个客户端同时对服务器发起访问
		Server: {"A": "1"}
		/      \
	   /        \
  client 1     client 2
```

考虑如下操作顺序：
```
# Step - 1: both clients 1 and 2 run Get: both fetch "A": "1"

# Step - 2: client 1 runs Put("A", "100"): server has {"A": "100"}

# Step - 3: client 2 runs Put("A", "999"): server has {"A": "999"}
```

在以上操作中，真正的问题不是服务器端最后的状态是 {"A": "999"}，而是client 2在仍旧认定A对应的value是"1"的情况下做出了将A的值修改为"999"的决定。然而实际情况是：此时A的value已经被client 1修改为"100"了，client 2在毫不知情的情况下，将A的值再次更改。

更为理想的情况应该是在第三步时，client 2得知A的值不再是自己在第一时间点获取的“1”时，重新获得A对应最新的value（是被client 1更新过的"100"），然后再进行更新，虽然最终服务器端key value store的状态并没有不同，但是client 2可以全程清晰的知道自己每一步操作时key-value的状态。

为了解决这个问题，我们给每一个key-value pair配一个对应的整型数version，并且Put操作也会传入key，value，version三个参数。只有当函数调用的version与服务器端的version相等时，Put操作才会执行更新key-value store，不然应该要拒绝修改。Get操作并不受引入version的影响（原因？“读”永远比“写”简单：“读”并不会修改数据而影响别的clients可能看到的数据）

引入version之后，我们再来看看上述的例子会如何变化：在最开始时：
```
# key: (value, version)
{"A": ("1", 0)}
```

再次考虑同样的操作：
```
# Step - 1: both clients 1 and 2 run Get("A"): 
#           both fetch "A": value="1", version=0

# Step - 2: client 1 runs Put("A", "100"， 0): 
            Put version matches server key value version: 0==0
	        update key-value while increasing version to represent edit times
	        {"A": ("100", 1)}  <-- Caution: version++

# Step - 3: client 2 runs Put("A", "999", 0):
			Put version is 0, but server version is 1
			Put request failed
			
# Step - 4: client 2 runs Get("A“)：
			fetch ("A": ("100", "1"))
			
# Step - 5: based on Step - 4 outcome, client 2 runs Put("A", "999", "1")
			Put succeeds
			server has {"A": ("999", 2)}
```

我们发现，在第二步时，当client 1成功执行Put操作后，服务器端key value store内的version比初始值增加1，表示该key-value已经被修改过1次。

在第三步client 2再次执行Put操作时，因为传入的version与服务器的version并不一致，导致client 2的Put操作失败。

因为Put操作失败，client 2此时可以再次执行Get操作，得知A对应的value是修改后的100，version为1，这是第四步。

紧接着，client 2可以再次执行Put操作，并且传入自己想修改的参数999以及version=1，服务器更新成功最终状态为`{"A": ("999", 2)}`

基于以上分析，我们可以设计类似`{key: (value, version)}`的数据结构，go语言中如下：

```go
type Entry struct {
	Value string
	Version rpc.Tversion
}

type KVServer struct {
	mu sync.Mutex

	// Your definitions here.
	table map[string]Entry
}
```

### Server::Get(Key)函数的实现

服务器端的Get函数实现相对直接，我们只需要考虑三点：
* 当一个客户端执行Get操作时，另一个客户恰好修改了key-value store，为了防止race condition，我们需要引入mutex。
* 从设计好的key-value store中根据key获取对应信息类似Python dict的操作。
* Get函数的返回不仅仅是value，还需要包括version，以及可能的error信息（客户端需要这些信息）。

基于以上考量，我们的Get函数实现如下：

```go
func (kv *KVServer) Get(args *rpc.GetArgs, reply *rpc.GetReply) {
	// lock to avoid race condition
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// directly fetch the key
	entry, ok := kv.table[args.Key]

	// key exist or not --> return value, version, error info
	if ok {
		reply.Value = entry.Value
		reply.Version = entry.Version
		reply.Err = rpc.OK
	} else {
		reply.Err = rpc.ErrNoKey
	}
}
```

### Server::Put(Key, Value, Version)函数的实现

服务器端的Put函数实现相比Get函数稍微复杂，我们需要考虑下面几点：
* 为了防止race condition，我们需要引入mutex。
* 当Put函数修改key对应的value时，需要区分key是否存在情况：
	* 如果key不存在时，我们需要观察Put函数传入的version：如果Version为0，则服务器创建对应的key-value pair。否则返回报错。（为什么version必须是0？因为一个全新的key-value pair未被任何客户端修改过，所以version必须是0）
	* 如果key存在时，我们仍需根据Put函数传入的version做出不同分支：如果传入的version与服务器存的version一致，则更新key-value pair并加version自增，否则返回报错。

整个服务器端的函数逻辑总结如下：
```
                    Put(key, value, version)
                             │
                             ▼
                 Does the key already exist?
                     ┌────────┴────────┐
                     │                 │
                    No                Yes
                     │                 │
         version == 0 ?       version == current?
             ┌───┴────┐          ┌────┴────┐
             │        │          │         │
            Yes      No         Yes       No
             │        │          │         │
             ▼        ▼          ▼         ▼
      Create new    ErrNoKey   Update    ErrVersion
      version = 1               value
                                version++
```

具体实现如下：

```go
func (kv *KVServer) Put(args *rpc.PutArgs, reply *rpc.PutReply) {
	// Your code here.
	// lock
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// attempt to update value
	// leverage golang map property like python dict.get()
	entry, ok := kv.table[args.Key]

	if !ok {
		if args.Version == 0 {
			kv.table[args.Key] = Entry{
				Value: args.Value,
				Version: args.Version + 1,
			}
			reply.Err = rpc.OK
		} else {
			reply.Err = rpc.ErrNoKey
		}

	} else {
		if args.Version == entry.Version {
			kv.table[args.Key] = Entry{
				Value: args.Value,
				Version: entry.Version + 1,
			}
			reply.Err = rpc.OK
		} else {
			reply.Err = rpc.ErrVersion
		}
	}

	return 
}
```
## 客户端的实现

客户端部分我们同样需要实现两个功能：Get和Put。只是在实现这两个功能时，我们需要考虑通过PRC远程调用服务器端的Get和Put。

因为网络通信的不稳定（分布式系统的一大难点），客户端需要在恰当的时机进行retry操作，直到收到确切的服务器端的回信。秉持着这一思考，我们分别对Get和Put函数做如下复盘：

### Client::Get函数

客户端的Get函数实现相对直观，我们只需要远程调用服务器端的Get函数，然后根据服务器返回的value，version，和error信息做出不同的操作：
* 如果error信息表明Get成功，我们只需要parse出value，version，error信息返回即可。
* 如果error信息表明key不存在，我们只需要返回value=empty string, version=0, error=no key

与之相对，如果客户端没有收到任何RPC调用结果，则需要让客户端不断retry，直到获得明确的服务器回信，然后回到上述分支进行操作。

此外，我们在实现客户端函数时，每次调用PRC实际上包含了两层函数的调用：
* RPC框架本身的调用：正常请求会返回 `bool true`。超时请求则会返回 `bool false`。这个调用结果仅仅表示“我们是否成功的调用了RPC”，而不代表具体执行服务器端相对应函数的结果。
* 服务器端定义的Get函数调用：服务器的Get函数会返回value，version，error，并将它们包装在 `reply GetReply`。这里的结果才是服务器端函数运行的结果。

举例来说，下面这个代码是客户端发送Get请求：
```go
ok := client.Call(ck.server, "KVServer.Get", &args, &reply)
```

变量`ok`仅仅表示RPC调用是否成功，而服务器函数Get的响应结果村在`reply`变量中。

根据以上逻辑，Get函数实现如下：
```go
func (ck *Clerk) Get(key string) (string, rpc.Tversion, rpc.Err) {
	// construct RPC request body
	args := rpc.GetArgs{
		Key: key,
	}
	for {
		// construct RPC reply body
		reply := rpc.GetReply{}
		
		// call RPC server::Get()
		ok := ck.clnt.Call(ck.server, "KVServer.Get", &args, &reply)
		if ok {
			// if RPC call is successful
			if reply.Err == rpc.OK {
				return reply.Value, reply.Version, reply.Err
			}
			if reply.Err == rpc.ErrNoKey {
				return "", 0, reply.Err
			}
		} else {
			// if RPC call fails due to network loss: retry
			time.Sleep(100 * time.Millisecond)
		}
	}
}
```

### Client::Put函数

客户端Put函数的实现是整个Lab最有挑战的部分，我们需要考虑的是：如果Put函数没有收到服务器回信，客户端能不能像Get函数一样“无脑”retry Put操作？

我们来看下面两个场景：

场景-1
```
# Time point 1
client sends Put("A", "100", 0)  # key, value, version

# Time point 2
server recevies Put and update key-value store: {"A": ("100", "1")}

# Time point 3
server reply gets lost;
client waits too long (timeout) and receive nothing
```

场景-2：
```
# Time point 1
client sends Put("A", "100", 0)  # key, value, version

# Time point 2
client sending message gets lost;
server receives nothing and does nothing to key-value store

# Time point 3
client waits too long (timeout) and receive nothing
```

在上述两种场景下，服务器端的key-value store发生了不同的变化，但是客户端的视角永远是timeout：RPC调用不成功。这也是分布式系统的一大难题：因为网络通信的不稳定，客户端远程调用失败时，永远无法准确的知道客户端一侧到底发生了什么。

基于以上的分析，我们来考虑两种情况下的Put操作：

**场景-1**
```
# server maintains map: {"A": ("1", 0)}  key-value-version

# client makes 1st attempt to call Put("A", "100", 10)

# server returns error version because 10 != 0
```


**场景-2**
```
# server maintains map: {"A": ("1", 0)}  key-value-version

# client makes 1st attempt to call Put("A", "100", 0)

# server receives the Put and update: {"A": ("100", 1)}

# server returns Put succeed message but get lost

# client waits for too long and makes 2nd attempt to call Put("A", "100", 0)

# server has {"A": ("100", 1)} and return version error because 1 != 0
```

仔细观察上述两个场景，client最终都会收到"error: version mismatch"的错误信息，但是场景-1下的"error: version mismatch"是可信的，因为这是客户端第一次发起请求，表明客户发送的Put函数里version是有问题的，因此Put操作绝对没有成功。

但是在场景-2里，客户第一次请求已经成功，但是服务器回信丢失，因此不得不的发送第二次请求，在第二次请求时，因为version不匹配而再次收到"error: version mismatch"的错误信息，这个错误信息并不像场景-1里的Put操作没有成功，而是Put操作可能在第一次成功（也可能没有成功）。

通过对比可以得知：
* 如果是客户首次发送Put请求而收到Error Version Mismatch错误，则肯定可以判定Put请求失败，失败原因是Put函数的Version有错。
* 如果是客户第二次或以后的Put请求收到Error Version Mismatch错误，这个错误既可以理解为上一次Put成功从而导致这次同样Put请求失败，也可以理解为Put操作的version参函数从始至终都是错的。

因此我们在实现Put函数时，需要考虑这是否是客户端第一次调用PRC（参照下面的代码更好理解）。

具体的Put实现如下：

```go
func (ck *Clerk) Put(key string, value string, version rpc.Tversion) rpc.Err {
	// construct RPC request body
	args := rpc.PutArgs{
		Key: key,
		Value: value,
		Version: version,
	}
	
	// whether this is initial Put or repeated Put call
	isFirstAttempt := true

	for {
		// construct RPC reply body
		reply := rpc.PutReply{}
		// make RPC call
		ok := ck.clnt.Call(ck.server, "KVServer.Put", &args, &reply)
		if ok {
			// if RPC call succeed
			// server Put() succeed
			if reply.Err == rpc.OK {
				return rpc.OK
			}
			// server Put() returns no key
			if reply.Err == rpc.ErrNoKey {
				return rpc.ErrNoKey
			}
			// server Put() returns error version but client 1st attempt
			if reply.Err == rpc.ErrVersion && isFirstAttempt {
				return rpc.ErrVersion
			}
			// server Put() returns error version and client >= 2nd attempts
			if reply.Err == rpc.ErrVersion && !isFirstAttempt {
				return rpc.ErrMaybe
			}
		} else {
			// RPC call fails: retry (so >=2nd time Put call)
			isFirstAttempt = false
			time.Sleep(100 * time.Millisecond)
		}
	}
}
```

## 分布式锁的实现

在以上的过程中，我们已经实现了version controlled key-value store，并且可以处理网络丢包的情况。在这一部分，我们可以利用已经实现好的key-value store来设计一个分布式锁。

尽管分布式锁与key-value store看上去大相径庭：锁是用来控制同一时间内只有某一段代码可以执行从而防止一侧读取一侧修改数据的情况；而key-value store则更多是用来储存数据信息。但其实我们可以利用key-value pair来表征锁：

比如我们的key就是锁的名字，它的value对应锁的状态或者锁的所有者ID：

```
# initial state:
"database_lock": ("unlock", 0)

# after client A acquires the lock
"database_lock": ("clientA", 1)

# after client A releases the lock
"database_lock": ("unblock", 2)
```

伴随着这样的理解，我们可以分成锁的设计，Acquire，以及Release来阐述

### 锁的数据结构

因为我们要用key-value pair来表示锁，自然而然我们可以想到我们的锁需要一个名字作为key，然后其value则是锁的状态：
* 当锁处于闲置状态时，value可以是"unblock"字样
* 当锁处于被获时，value则是锁持有者的unique identity

基于以上考量，我们设计如下数据结构来表示锁：

```go
const (
	// lock state: unblock
	UNLOCK = "UNLOCK"
	// we do not need "lock" because when it is locked
	// it will hold owner's name rather than simply writing "lock"
)

type Lock struct {
	// IKVClerk is a go interface for k/v clerks: the interface hides
	// the specific Clerk type of ck but promises that ck supports
	// Put and Get.  The tester passes the clerk in when calling
	// MakeLock().
	ck kvtest.IKVClerk
	// You may add code here
	lockName string
	clientID string  // unique identity for each client
}
```

### Acquire功能的实现

获取锁的逻辑如下图所示：我们需要利用Get来获取当前锁的状态，并在合适的时候利用Put函数获取锁。

具体的逻辑阐述如下：
* 首先调用Get函数，获取锁名作为key对应的value与version：如果value恰好等于自己ID，说明自己已经获取了锁。如果锁名是别的客户ID，则直接等待并不停尝试。
* 当Get函数发现锁的状态是闲置时，我们可以尝试调用Put函数来获取锁，Put函数会传入锁名（key），自己的ID（value），对应的version（从上一步Get函数获得）。这里通过version可以有效的防止两个客户同时获得锁：
```
Client A: 
Get(lock) -> UNLOCK, version = 5

Client B:
Get(lock) -> UNLOCK, version = 5

Both clients send Put(lock, ID, 5): Only one succeed and version will be come 6. The other client will fail.
```

另外值得注意的是，当Put函数返回Error May Be时，这是一个模棱两可的错误信息，它即可表示前一次获取锁的Put请求已经成功，但是网络通讯丢失，所以重新尝试第二次Put，也可以意味着本身Put请求的version就有错。在这个分支下，我们需要再次调用Get函数来确认锁的持有者是否是自己，如果是自己，则说明我们上一次Put已经成功，只是没有收到回信。

具体的逻辑简图如下：
```
Acquire()
    ↓
Get(lockName)
    ↓
Is the lock available?
    ↓
Yes ──► Put(lockName, clientID, version)
 |          |
 |          ├── OK          → Lock acquired
 |          ├── ErrVersion  → Another client won the race
 |          └── ErrMaybe    → Verify ownership
 |
No
 ↓
Retry
```

函数实现如下：
```go
func (lk *Lock) Acquire() {
	// Your code here
	for {
		// get current state
		value, version, _ := lk.ck.Get(lk.lockName)

		if value == lk.clientID {  // if hold the lock and owner is me
			return
		}

		// if locked but owner is not me, do nothing but to wait
		if value != "" && value != UNLOCK && value != lk.clientID {
			continue
		}
		
		// all other cases: attempt to lock --> put my ID there as a lock state so that others know it's me
		putErr := lk.ck.Put(lk.lockName, lk.clientID, version)

		// if success return
		if putErr == rpc.OK {
			return
		}
		if putErr == rpc.ErrMaybe {  // if error case: confirm whether I have already gained the lock
			value, _, _ := lk.ck.Get(lk.lockName)
			if value == lk.clientID {
				return
			}
		}
	}
}
```

### Release功能的实现

锁的释放功能实现相对直观：我们先用Get函数来获取锁对应的状态（value），然后根据不同情况来决定是否调用Put函数来释放锁。

我们同样需要考虑，如果通信丢失时，需要再次调用Get函数来确认自己是否已经成功的释放锁。

函数实现如下：
```go
func (lk *Lock) Release() {
	// Your code here
	for {
		value, version, _ := lk.ck.Get(lk.lockName)
		if value == UNLOCK {  // already released: nothing to do
			return
		}
		putErr := lk.ck.Put(lk.lockName, UNLOCK, version)
		if putErr == rpc.OK {
			return
		}
		if putErr == rpc.ErrMaybe {
			value, _, _ = lk.ck.Get(lk.lockName)
			if value == UNLOCK {
				return
			}
		}
	}
}
```

# Lab小结

通过这次Lab，可以发现：
* 即便是简单的key-value store也可以实现分布式锁：Redis作为锁的原理跟这个应该相似
* 分布式系统中一大难题就是当网络丢失时，客户端永远不知道服务器端发生了什么，因此需要设计不同的错误信息来消除误解
* 引入version number可以非常轻量级的实现分布式系统各个节点协调的功能
