[TOC]

# Transfer requests

This package is implementation of transfer requests, it's built based on the "transfer" package.
See [transfer and transfer builder](../../../../internal/transfer/readme.md)

## General information

Although the files contain a lot of source code, the code is essentially simple to understand.
A transfer creation consists of the following steps:

* Formation of a request for transfer and initialization of incoming data (inputs);
* Transfer modeling (see private method "evaluate");
* Checking permissions;
* Saving results in the database.

### Request formation

See `internal/modules/request/creator.go` which is responsible for request creation.
It will be refactored in near future.

### Transfer modeling 

As you can see different types of transfers requires different input data:

* BA(Between Accounts) requires a source and destination accounts;
* CFT(Card Funding Transfer) requires a source account and destination card;
* DA(Debit Account) requires only source account;
* CA(Credit account) requires only destination account;
* DRA(Deduct revenue accounts) requires revenue account;
* OWT(Outgoing wire transfer) requires source account.

Also, different types of transfer may require their specific parameters such as "fee parameters", or some
options that could be applied to the transfer.
 
Although input data is different all the transfers come through the same steps - a sequence of actions is formed using
[the transfer](../../../../internal/transfer/readme.md) package. The input data affects the formation of this sequence.

There are only 3 kinds of the actions which completely define any transfer.

* Debit;
* Credit;
* Exchange.

Only **debit** and **credit** actions spawn a transaction.
**Exchange** is an auxiliary action that calculates values in different currencies.

The most common pattern you can see in the "evaluate" is something similar to:

```go
	chain.
        // debit some sum
		Debit(sum).
        // from the source (account)
		From(source).
        // callback is used in order to retrieve required data that are used
        // create a transaction model and fill in the details
		WithCallback(func(action transfer.Action) error {
			err := action.Perform()
			currency := action.Currency()
            // create transaction model
			transaction := &txModel.Transaction{
				// ...
			}
            // add the transaction to the list for later use (save/update DB)
			b.appendTransaction(transaction)
            // the details of the transfer reflect the sequence of actions and can be used 
            // for a wide variety of tasks such as 
            // "notifications", "displaying data to the user", "permission checks", "validation" etc.
			details[debitPurpose] = &types.Detail{
                //...
			}
			return err
		})
```

#### **DryRun** vs **Evaluate**

You can find such methods in transfer handlers. Essentially, they do the same thing - modeling
transfer. The difference is that **DryRun** does not change model balances (accounts, card, revenue account). At that
while **Evaluate** changes.

### Transfer permissions

Permissions **ARE checked only before creating a new transfer**. To evaluate, modify or execute an existing transfer
permissions **are NOT validated**.

Any transfer permission is implemented with the below interface:
```go
// PermissionChecker defines the contract that declares how to check if transfer is allowed
type PermissionChecker interface {
	// Check checks whether transfer is allowed
	Check() error
	// Name returns permission specific name
	Name() string
}
```

Currently, **permissions are not applied** to "CA", "DA", "DRA" transfers.

#### Default permissions

* Withdrawal allowed - applied to source accounts. Defined by the "AllowWithdrawals" account model field.
* Deposit allowed - applied to destination accounts. Defined by the "AllowDeposits" account model field.
* Sufficient Balance - applied to source account. Checks whether account available balance is enough for the transfer.

#### Custom permissions

Transfer constructors accept "PermissionFactory". To set custom permissions, you can replace default factory with the 
custom "PermissionFactory".

```go
// PermissionFactory is used in order to define permissions
type PermissionFactory interface {
	CreatePermission(request *requestModel.Request, details types.Details) (PermissionChecker, error)
	WrapContext(db *gorm.DB) PermissionFactory
}
```

#### "Limit" permissions

> Limits could be managed with RPC, see [limit server examples.](../../../../internal/limitserver/readme.md)

There are several permissions that impose certain limitations on the various values involved in transfers. 
Using these **limitations requires significant system resources** and significantly slows down the operation of transfers. 
**By default**, all limitations are **enabled**. 
These limitations (all or some) **can be disabled**. 
In order to disable the "limit" permission, it is necessary to redefine the corresponding 
constant which looks like this `Limit{limit name}Enabled`, see `limit * .go` files.

Also, using constants, you can customize the default limit values. For more details about the implementation see
[the limit](../../../../internal/limit/readme.md) package description.

The limits are indicated in a specific currency. 
For calculations, the amounts in different currencies are converted to the limit currency. 
The current exchange rates(currencies service) at the time of calculation are used.

**LimitMaxTotalBalance** 

This permission imposes limitation on the total balance of all user accounts. 
Transfers in "pending" status are also included in the calculation.

Impact on performance: high.

`(A + P) < M`

* A - sum of all user account available balances
* P - absolute sum of all pending transactions related to a user accounts
* M - value of the "LimitMaxTotalBalance" permission


**LimitMaxCreditPerTransfer**

This permission limits the maximum amount that can be credited to a user's account within a single transfer.

Impact on performance: low.

`A < M`

* A - sum of all credit transactions related to a user account within a single transfer.
* M - value of the "LimitMaxCreditPerTransfer" permission.

**LimitMaxDebitPerTransfer**

This permission limits the maximum amount that can be debited from a user's account within a single transfer.

Impact on performance: low.

`A < M`

* A - absolute sum of all debit transactions related to a user account within a single transfer.
* M - value of the "LimitMaxDebitPerTransfer" permission.


**LimitMaxTotalDebitPerDay** and **LimitMaxTotalDebitPerMonth**

These permissions limit the maximum amount that can be debited in total from all user accounts 
within a given time period. 
Transfers in "pending" and "executed" statuses are used.

Impact on performance: high.

In case of "LimitMaxTotalDebitPerDay" the period is current day starting from 00:00:00 till 23:59:59 of the current day.

In case of "LimitMaxTotalDebitPerMonth" the period is the first day of the current month starting from 00:00:00 
till 23:59:59 of the last day of the current month.

`(Ap + Ae) < M`

* Ap - absolute sum of all transactions in "pending" state related to a user accounts per defined period.
* Ae - absolute sum of all transactions in "executed" state related to a user accounts per defined period.
* M - value of the corresponding limit permission.