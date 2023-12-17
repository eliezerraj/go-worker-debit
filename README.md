# go-worker-debit

POC for test purposes.

Worker consumer kafka topics from the go-fund-transfer service

## diagram

kafka <==(topic.debit)==> go-worker-debit (GROUP-02) (post:/add) ==(REST)==> go-debit(Service.Add)
