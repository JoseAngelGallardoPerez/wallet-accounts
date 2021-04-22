[TOC]

# Limit server

Limit server implements the RPC server, which provides the ability to set transfer limits.

## Related content (highly recommended)

* [Limit](../limit/readme.md)
* [Transfers](../modules/request/transfers/readme.md)

## Init client

```go
client := rpcLimit.NewLimitsProtobufClient("http://service-rpc.addr.", http.DefaultClient)
```

## Set limit

```go
    // the request below creates/updates 3 limits at once
    // max total balance is 175 BTC
    // no limit for "max debit per transfer"
    // 7500 AUD for "max credit per transfer"
    // all for user with id "user_1"
	_, err := client.Set(context.Background(), &rpcLimit.SetLimitsRequest{
		Limits: []*rpcLimit.LimitWithId{
			{
				LimitId: &rpcLimit.LimitId{
					Name:     rpcLimit.LimitName_MAX_TOTAL_BALANCE,
					Entity:   "user",
					EntityId: "user_1",
				},
				Limit: &rpcLimit.Limit{
					CurrencyCode: "BTC",
					Amount:       "175",
				},
			},
			{
				LimitId: &rpcLimit.LimitId{
					Name:     rpcLimit.LimitName_MAX_DEBIT_PER_TRANSFER,
					Entity:   "user",
					EntityId: "user_1",
				},
				Limit: &rpcLimit.Limit{
					NoLimit: true,
				},
			},
			{
				LimitId: &rpcLimit.LimitId{
					Name:     rpcLimit.LimitName_MAX_CREDIT_PER_TRANSFER,
					Entity:   "user",
					EntityId: "user_1",
				},
				Limit: &rpcLimit.Limit{
					CurrencyCode: "AUD",
					Amount: "7500",
				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("ok")
```

## Get limit

**You could specify as many limit ids as you need**

```go
	response, err := client.Get(context.Background(), &rpcLimit.GetLimitsRequest{
		Identifiers: []*rpcLimit.LimitId{
			{
				Entity:   "user",
				EntityId: "user_1",
			},
			// this one is not exist
			{
				Name: rpcLimit.LimitName_MAX_TOTAL_DEBIT_PER_DAY,
				Entity:   "user",
				EntityId: "user_1",
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	limits := response.Limits

	for _, limitData := range limits {
		lim := limitData.Limit
		id := limitData.LimitId

		fmt.Printf("Limit Name: %s, Entity: %s, EntityId: %s\n", id.Name, id.Entity, id.EntityId)
		if !lim.Exists {
			fmt.Println("Limit is not exist")
		}
		if lim.NoLimit {
			fmt.Println("No Limit")
		} else {
			fmt.Printf("Amount: %s %s\n", lim.Amount, lim.CurrencyCode)
		}
		fmt.Println("--------------------------------")
	}
```

Will print:

```
Limit Name: MAX_TOTAL_BALANCE, Entity: user, EntityId: user_1
Amount: 175 BTC
--------------------------------
Limit Name: MAX_DEBIT_PER_TRANSFER, Entity: user, EntityId: user_1
No Limit
--------------------------------
Limit Name: MAX_CREDIT_PER_TRANSFER, Entity: user, EntityId: user_1
Amount: 7500 AUD
--------------------------------
Limit Name: MAX_TOTAL_DEBIT_PER_DAY, Entity: user, EntityId: user_1
Limit is not exist
```

## Reset/remove existing limits (set to default)

```go
	_, err := client.ResetToDefault(context.Background(), &rpcLimit.ResetLimitsRequest{
		Identifiers: []*rpcLimit.LimitId{
			{
				Name:     rpcLimit.LimitName_MAX_TOTAL_BALANCE,
				Entity:   "user",
				EntityId: "user_1",
			},
			{
				Name:     rpcLimit.LimitName_MAX_DEBIT_PER_TRANSFER,
				Entity:   "user",
				EntityId: "user_1",
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("ok!")
```