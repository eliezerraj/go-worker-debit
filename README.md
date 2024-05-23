# go-worker-debit

POC for test purposes.

Worker consumer kafka topics from the go-fund-transfer service

## diagram

kafka <==(topic.debit)==> go-worker-debit (GROUP-02) (post:/add) ==(REST)==> go-debit(Service.Add).

## Database

    CREATE TABLE public.transfer_moviment (
        id serial4 NOT NULL,
        fk_account_id_from int4 NULL,
        fk_account_id_to int4 NULL,
        type_charge varchar(200) NULL,
        transfer_at timestamptz NULL,
        currency varchar(10) NULL,
        amount float8 NULL,
        status varchar(200) NULL,
        CONSTRAINT transfer_moviment_pkey PRIMARY KEY (id)
    );

    ALTER TABLE public.transfer_moviment 
    ADD CONSTRAINT transfer_moviment_fk_account_id_from_fkey 
    FOREIGN KEY (fk_account_id_from) REFERENCES public.account(id);
    ALTER TABLE public.transfer_moviment 
    ADD CONSTRAINT transfer_moviment_fk_account_id_to_fkey 
    FOREIGN KEY (fk_account_id_to) REFERENCES public.account(id);
