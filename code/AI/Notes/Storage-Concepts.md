# Storage System Design - Interview Quick Answers

## Q1. Explain Block, File, and Object Storage.

- **Block Storage:** Stores data as fixed-size blocks and is ideal for databases, VMs, and boot volumes. Provides low latency and random I/O.
- **File Storage:** Organizes data in a hierarchical file system with directories and file metadata. Suitable for shared file access (NFS, SMB).
- **Object Storage:** Stores data as objects containing data, metadata, and a unique ID. Highly scalable and ideal for backups, media, and AI datasets.

---

## Q2. Why is object storage more scalable than file storage?

- Object storage uses a flat namespace and distributes objects across multiple nodes, avoiding metadata bottlenecks.
- It can scale to billions of objects, whereas file storage is limited by filesystem metadata management.

---

## Q3. Explain RAID.

- RAID (Redundant Array of Independent Disks) combines multiple disks to improve performance, capacity, or fault tolerance.
- Common levels: RAID 0 (performance), RAID 1 (mirroring), RAID 5 (single parity), RAID 6 (dual parity), RAID 10 (performance + redundancy).

---

## Q4. What is NVMe?

- NVMe (Non-Volatile Memory Express) is a storage protocol designed for SSDs over PCIe.
- It supports multiple queues, high parallelism, and low latency, making it much faster than SATA/SAS.

---

## Q5. What is NVMe-oF?

- NVMe over Fabrics extends the NVMe protocol across a network using RDMA, TCP, or Fibre Channel.
- It enables remote SSDs to provide performance close to locally attached NVMe devices.

---

## Q6. Explain Ceph architecture.

- Ceph consists of **MON** (cluster state), **OSD** (stores data), **MGR** (monitoring/management), and **MDS** (CephFS metadata).
- Clients use the **CRUSH** algorithm to locate data directly without a central metadata server.

---

## Q7. Why doesn't Ceph use a metadata server for block objects?

- Ceph uses the deterministic **CRUSH** algorithm to calculate object placement, eliminating metadata lookups.
- This removes a central bottleneck and improves scalability and availability.

---

## Q8. Explain CRUSH.

- CRUSH (Controlled Replication Under Scalable Hashing) is Ceph's placement algorithm that maps objects to OSDs.
- It is topology-aware and ensures replicas are distributed across different failure domains.

---

## Q9. Explain Consistent Hashing.

- Consistent hashing distributes data across nodes while minimizing data movement when nodes are added or removed.
- It is widely used in distributed caches, databases, and storage systems.

---

## Q10. Difference between CRUSH and Consistent Hashing.

- **Consistent Hashing** distributes data across nodes but has no awareness of physical topology.
- **CRUSH** is topology-aware, understands racks/hosts, and places replicas intelligently for fault tolerance.

---

## Q11. Explain Sharding.

- Sharding partitions data across multiple servers to improve scalability and throughput.
- Common strategies include **range-based**, **hash-based**, and **directory-based** sharding.

---

## Q12. Can sharding and consistent hashing be used together?

- Yes. Sharding determines which partition stores the data, while consistent hashing distributes each shard across multiple nodes.
- Many distributed databases combine both techniques for scalability and load balancing.

---

## Q13. Explain the complete write path.

- Client → Storage API → Metadata/Placement → Primary Storage Node → Replication/Erasure Coding → Persistent Storage → Acknowledgment.
- The write is acknowledged only after meeting the configured durability requirements.

---

## Q15. Why separate control plane and data path?

- The **control plane** manages metadata, configuration, scheduling, and policies.
- The **data path** handles actual I/O operations, allowing both to scale and evolve independently.

---

## Q18. Explain checksum.

- A checksum is a small value computed from data to detect accidental corruption during storage or transmission.
- Storage systems use checksums to detect bit rot and data corruption.

---

## Q19. Difference between checksum and hash.

- **Checksum** detects accidental errors and is optimized for speed.
- **Hash** uniquely identifies data with strong collision resistance and is used for integrity verification and deduplication.

---

## Q20. Explain deduplication.

- Deduplication eliminates duplicate data by storing only one copy and referencing it multiple times.
- It reduces storage capacity requirements, especially for backups and VM images.

---

## Q21. Explain compression.

- Compression reduces the size of data by encoding it more efficiently.
- It saves storage space and reduces network bandwidth while requiring CPU resources for compression/decompression.

---

## Q22. Why compression before erasure coding?

- Compressing first reduces the amount of data that must be encoded.
- This lowers storage consumption, network traffic, and erasure coding overhead.

---

## Q23. Explain erasure coding.

- Erasure coding splits data into **k** data chunks and **m** parity chunks, allowing recovery from failures.
- It provides fault tolerance with much lower storage overhead than full replication.

---

## Q25. Explain write amplification.

- Write amplification occurs when the storage system writes more physical data than the application requested.
- It is caused by garbage collection, wear leveling, journaling, and SSD internal management.

---

## Q29. Explain GPUDirect Storage.

- GPUDirect Storage allows storage devices to transfer data directly into GPU memory, bypassing CPU and system RAM.
- This reduces CPU overhead and improves AI training and inference throughput.

---

## Q30. Explain RDMA.

- RDMA (Remote Direct Memory Access) enables direct memory-to-memory communication between servers without involving the remote CPU.
- It provides low latency, high throughput, and is widely used in AI, HPC, and NVMe-oF.

---

## Q31. Explain Quorum.

- Quorum is the minimum number of nodes required to make cluster decisions safely.
- It prevents split-brain scenarios and ensures data consistency during failures.

---

## Q36. Why do enterprise SSDs have power-loss protection?

- Enterprise SSDs use capacitors to flush in-flight writes from DRAM to NAND during unexpected power failures.
- This prevents data corruption and ensures write durability.

---

## Q50. Explain how you would debug: "Writes are slow."

- Check **application latency**, **disk utilization**, **IOPS**, **bandwidth**, **CPU**, **memory**, **network**, and **storage health**. Identify whether the bottleneck is in the client, network, storage nodes, or disks.
- Review logs, replication/erasure coding delays, queue depths, latency metrics, and recent configuration or hardware changes to isolate the root cause.


Since the role involves Go and storage systems, you may also be asked to implement:
LRU Cache
Thread-safe Cache
Consistent Hashing
Producer–Consumer
Worker Pool
Read/Write Lock
Rate Limiter
File Chunker
Merkle Tree
Checksum Validator