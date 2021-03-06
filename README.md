Grex
====

Grex - a flock of storage - and application server cluster - sharded and redundant - with apllication object cache coordination

Why
---
To store millions (or trillions) of records reliably and fast on small machines that dont have Gigs of RAM and Terabyte disks

Most existing key value stores need loads of RAM to operate effectively, and themselves have large memory footprints - for example Cassandra

To take advantage of optimisations in Go particularly its own improvement over protocol buffers - the mighty GOB

To include the cluster code in the app servers to cache at the decoded object (application level)

To take advantage of changes in computing and networking and comparative performnace.

Design Principle
----------------

Clusterable by default - for replication and sharding

No single point of failure - identical peer nodes

Flexible persistance back end - (uses native file system by default, with no compression, but it would be trivial to plug in snappy for example), but could be anything - add a backend that pipes data to S3 or Appengine for example.

Bucket based locality of reference advantages.
  
Data stored in buckets based on primary bucket key, and item sub-keys within a bucket - for example a user id would be the bucket, and their list of favourite web sites would be encoded into an item

Each item within a bucket is just binary data - a (byte array) - but its expected that your app will gob encode and decode the binary data.

Groups of buckets are organised into 'flocks' by some computed grouping - a flock is the smallest division that a node can redistribute round the cluster - i was using first two characters of the bucket key - but (using the snapchat usernames as test data) the distribution sucked so now its 10 bits of golangs inbuilt fnv1a hash - much better balanced

A node will be responsible for its copy of a flock, and may herd all flocks or a subset - the set of flocks that a node is responsible for (herding) is called its 'farm'

Replication and sharding are treated the same - all nodes know what flocks they have in their farm - and know enough detail of all other nodes farms to route requests to the best node.

Binary persistance and gob based RPC communication (bleeting) across the cluster - your application layer does the marshaling to and from user specific go types.

The replication and sharding algorithms is .... none existant you tell each node what it should be doing :)


When to use Grex
----------------

Although there is nothing to stop grex scaling infinitely - it is particulalry usefull when you have cheap small servers, and still want great perfomance.

When your data can be organised by a primary key which identifies buckets of data items where each item is normally bigger than the file system page size (4k - 16k). (blobs, or )

When you dont want to query the data - just get it back load it into memory - query / manipulate it with your own app code and store it in the cluster again.

When you can typically fit one bucket in memory.

When the inodes available to you are comparable to the number of buckets and items you will store per node.

When you want a Go library to include in your app server to turn your app server into a distributed cluster of app servers.

Typical use case
----------------

Your app is customer centric - a cutomer has a few record types eg contacts, orders, favourites etc.

The bucket key will be the customer reference, and the item sub key would be the record type:

  "johndoe", "contacts"
  "johndoe", "orders"
  "johndoe", "favourites"


Design
------
*consistancy....*

Lamport timestamps or Vector clocks and whatnot too complex and an uneccessary overhead for the app server scenario, given a couple of assumptions...

In an app server each request can be statefull - in that it has a session - linked to a specific bucket. So we can decide in the app layer that only one client can iniitiate changes to a bucket - if a client request causes a sequence of bucket updates on the server these could be synchronised within the cluster with vector clocks, or we can simply acquire a replication lock on the bucket for the duration of the server task.

In our case changes to a bucket are generally going to be via human input, or via a batch job - the human input can be controlled as above, and the batch jobs can aquire a lock when there is no bucket session.

Of course we could allow more than one session per bucket and the server can then deal with the vecor clocks / conflict resolution.

Eventual consistancy
--------------------

As writes to a particular replicated bucket can be sunchronised there is no need to care to much about temporal conflicts.

Consistancy is really a major issue in the case of node failure - by keeping a tiny bit of meta data for the bucket - recording the cluster version of each file - 

We have borrowed heavily from Riak (when they say vector clock read version)...

>How is divergence addressed? When you make a read request, Riak looks up all replicas for that object. By default, Riak will return the most updated version, determined by looking at the object's vector clock. Vector clocks are metadata attached to each replica when it is created. They are extended each time a replica is updated to keep track of versions. You can also allow clients to resolve conflicts themselves.

*Read Repair*

>Further, when an outdated replica is returned as part of a read request, Riak will automatically update the out-of-sync replica to make it consistent. Read repair, a self-healing property of the database, will even update a replica that returns a “not_found” in the event that a node loses it due to physical failure.



New job

 bucket key, jobname data

 look in cluster for anyone with the bucket in cache

 send job to them

 if job updated data then send binary data to all servers who care about this flock

 return url of active server for this bucket - including 'canary requests' to ask for best node



Influence
---------
The primary influence for Grex was Riak - 
http://basho.com/riak/

The LRU Cache comes from googles Vitesse - 
https://github.com/youtube/vitess

Are networks now faster than disks? - 
http://serverfault.com/questions/238417/are-networks-now-faster-than-disks

Dr Jeff Dean Keynote PDF - 
http://www.cs.cornell.edu/projects/ladis2009/talks/dean-keynote-ladis2009.pdf

MongoDB - 
http://docs.mongodb.org/manual/core/sharding/

Amazon S3 Principles - 
http://s3.amazonaws.com/doc/s3-developer-guide/OverviewDesignPrinciple.html



Some performance observations
-----------------------------

10 user accounts each with a list of 20,000 records like this...
	
	x1 := Xact{
		Description: "My very first description",
		Other:       "My other description",
		Amount:      1245,
		Date:        time.Now(),
	}

Each file is uncompressed takes up 1.4M meaning that each record (row) takes up about 70 bytes


Three nodes on one mac book pro computer wiht SSD all replicating - into different root folders.


One node generates the transactions and then within a loop encodes the list into a gob and saves them to all three nodes

it was able to loop at 4.2 loops per second....

4.2 * 10 * 20,000 = 840,000 records per second

replicated 3 times
4.2 * 10 * 1.4M * 3 = 176Mbps -> which is pretty typical of the SSD macbook 


Attempting to increase the loop has no effect on tese numbers - it maxes out at 4.2 per second.

The rpc connection pond (pool) only ever sees upto 8 out of the 20 availble connections per node being made - so its safe to assume that the local loopback is transferring data quicker than the disk can write it, which although obvious, also confirms that the overhead in encoding and calling the rpc's is not the bottleneck.

Having said that when the disk persistance is turned off the the loop only runs a little bit quicker

So I pre allocated the buffer being passed to the encoder - but that made little difference so...

After compiling with pprof I got this...

	(pprof) top10 -cum
	Total: 299 samples
	     165  55.2%  55.2%      276  92.3% time.Parse
	       0   0.0%  55.2%      261  87.3% runtime.(*errorString).Error
	       4   1.3%  56.5%       64  21.4% reflect.Value.Interface
	       5   1.7%  58.2%       61  20.4% reflect.valueInterface
	       0   0.0%  58.2%       47  15.7% strconv.Unquote
	      11   3.7%  61.9%       43  14.4% strconv.IsPrint
	       0   0.0%  61.9%       37  12.4% time.(*Time).GobEncode
	       7   2.3%  64.2%       37  12.4% time.Time.GobEncode
	       0   0.0%  64.2%       30  10.0% runtime.gc
	      13   4.3%  68.6%       23   7.7% runtime.FixAlloc_Free
	(pprof)

Which shows that a lot of time was spent parsing time

Then after removing the time type from the record the loop went up to 10 persecond with the profiler showing most of the time taken up encoding unicode - I stopped optimising then.

My gut instinct is that gob encoding can be made quicker with specifc (unsafe) methods for each known type - especially if you are rpc-ing between homogeneous nodes 













