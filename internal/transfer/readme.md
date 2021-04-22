# Transfer package

The main purpose of this package is to provide tools for modeling and calculations related to the 
transfer of funds.
One of the most important advantages of this package is its simplicity. 
There is no reflection, complex logical branches and side effects. 
All functionality is built on the basis of the simplest blocks.

This document explains concepts and their application to their practical meaning.

## The basic types

There are several simple types that are used in order to compose complex logic.
Please familiarize yourself with these types. 
In what follows, it will be demonstrated how these types are combined in order to implement the calculations.

**Balance** is used in order to determine a certain value and be able to change this value up or down.


```go
// Balance defines base operations
type Balance interface {
	// Amount shows current balance amount
	Amount() decimal.Decimal
	// Add adds value "v" to current balance
	Add(v decimal.Decimal) error
	// Sub subtracts value "v" from current balance
	Sub(v decimal.Decimal) error
}
```

**Currency** is used in order to define a specific currency and its divisibility.


```go
// Currency represents money currency information required for transfers
type Currency struct {
	code     string
	fraction uint
}

// Fraction represents a fraction of a currency unit (number of decimal places) e.g. 2 for USD - 10^2 (100 cents)
func (c *Currency) Fraction() uint {
	return c.fraction
}

// Code returns currency code i.e. EUR, USD etc.
func (c *Currency) Code() string {
	return c.code
}
```

**Amountable** is anything that can measured as "amount" and provide it.


```go
// Amountable represents types that could provide amounts
type Amountable interface {
	// Amount returns value representation
	Amount() decimal.Decimal
}
```

**CurrencyAmount** is used in order to express a certain value in a certain currency.


```go
// CurrencyValue represents types that contain both currency and amount
type CurrencyAmount interface {
	// Currency returns currency
	Currency() Currency
	// Amount returns value representation
	Amount() decimal.Decimal
}
```

**Debitable** is used in order to "debit" CurrencyAmount from something that implements this interface.


```go
// Debitable represents an instance that could be debited
type Debitable interface {
	CurrencyAmount
	// Debit debits funds from the debitable
	Debit(amount CurrencyAmount) error
}
```

**Creditable** is used in order to "credit" CurrencyAmount to something that implements this interface.


```go
// Creditable represents an instance that cold be credited
type Creditable interface {
	// Currency returns currency
	Currency() Currency
	// CreditFromAlias credits funds to the creditable
	Credit(amount CurrencyAmount) error
}
```

> I know, it can be tedious to look at so many types in a row, but please be patient, there are only a few left and we 
> will move on to practice examples.

**Performer** interface is used to encapsulate logic. 
It provides 2 methods: "Perform" to perform an action and "IsPerformed" to check if an action has been performed.


```go
type Performer interface {
	// Perform executes expected logic
	Perform() error
	// IsPerformed indicates whether action is performed
	IsPerformed() bool
}
```

**Action** is combination of "Performer" and "CurrencyAmount" interfaces extended with methods that allows identifying action
and determine result "Sign". This type is very important. In order to better understand the meaning of this type simply prefix
it with the package name `transfer.Action`. So the transfer action is a logical step i.e. credit/debit/exchange etc.

```go
// Action describes required operations that are used in order to create transfer action
type Action interface {
	CurrencyAmount
	Performer
	// Purpose describes action purpose
	Purpose() string
	// Message is optional piece of information that clarifies the action
	Message() string
	// Sign indicates whether action is debit (-1) or credit (1), (0) if action is not credit or debit
	Sign() int
}
```


## Let's do the simplest transfer

The simplest transfer is to move funds from one balance to another. Let's imagine that we have 2 wallets "A" and "B".
Wallet "A" has balance equal to 100, wallet "B" has zero balance.
As was stated above we have "Balance" interface that specifies all methods that we need in order to perform the transfer A -> B.

```go
// Balance defines base operations
type Balance interface {
	// Amount shows current balance amount
	Amount() decimal.Decimal
	// Add adds value "v" to current balance
	Add(v decimal.Decimal) error
	// Sub subtracts value "v" from current balance
	Sub(v decimal.Decimal) error
}
```

Here is how it may look like.


```go
var A, B transfer.Balance
// ...
fmt.Println(A.Amount()) // 100
fmt.Println(B.Amount()) // 0

// transfer
amount := decimal.NewFromInt(70)

// For the sake of example simplicity let's omit error checks
A.Sub(amount) // take 70 from A
B.Add(amount) // put 70 to B

// Now we can check the result
fmt.Println(A.Amount()) // 30
fmt.Println(B.Amount()) // 70
```

Simple enough in my opinion. Let's move on. Usually the transfer is carried out in some kind of currency.
Let's rewrite the example above, but this time we use types that take amount in certain currency.


```go
var (
    A transfer.Debitable
    B transfer.Creditable
    amount transfer.CurrencyAmount
)

// ... 
A.Debit(amount)
B.Credit(amount)
```

As you can see, the principle has not changed, but what does the use of the value in relation to the currency give us?
And how do we check the end result?
Using currency bound to a value allows us to check that we are adding or subtracting a value in
the same currency.
In order to check the result, we can go beyond implementing only interface methods.
The transfer package already contains the required *Wallet*  implementation. This type implements *Debitable*, *Creditable* and
**CurrencyAmount**, see `internal/transfer/wallet.go`


```go
var (
    A transfer.Debitable
    B transfer.Creditable
    amount transfer.CurrencyAmount
)
//..
euro := transfer.NewCurrency("EUR", 2)
amount = transfer.NewAmount(euro, decimal.NewFromInt(70))

A = transfer.NewWallet(transfer.NewSimpleBalance(decimal.NewFromInt(100)), *euro)
B = transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *euro)

// ... 
// Now if we try to debit amount in different currency then we will get an error
err := A.Debit(amount)
if err != nil {
    //...
}
// same here
err = B.Credit(amount)

// Now we can check the result
fmt.Println(A.Amount()) // 30
fmt.Println(B.Amount()) // 70

```
It is important to note that "Debit" and "Credit" can only accept values greater than zero. 
This restriction avoids undesirable behavior.
Since negative values can invert credit and debit operations. 
Developers must explicitly indicate the intention to perform debit or credit action by calling the corresponding methods.

The above example is also quite simple. As you may have noticed, the parameter that the "Debit" and "Credit" methods take is
an interface, this opens up the possibility for the implementation of calculations.

Let's complicate the example a little. Say now we need to calculate the amount that is debited from wallet "A" and credited
to wallet "B". Calculation still simple - 10% of a given amount.


```go
func DoTransfer(A transfer.Debitable, B transfer.Creditable, amount transfer.CurrencyAmount) {
    factor := decimal.NewFromFloat(0.1)
    // we could calculate 10% and create new amount like so
    // tenPercentAmount := transfer.NewAmount(amount.Currency(), amount.Amount().Mul(factor))
    // but there is a type that helps to do the same 
    tenPercentAmount := transfer.NewAmountMultiplier(amount, factor)
    A.Debit(tenPercentAmount)
    B.Credit(tenPercentAmount)
}
```

It is often necessary to perform an action depending on the conditions. For this, as well as for the formation of more complex
structures package provides implementation of the most basic actions that are used in transfers: "DebitAction", "CreditAction", and
"ExchangeAction" (we'll talk about it later).
"Actions" also brings more flexibility to the code. 


```go
    debit, _ := transfer.NewDebitAction(debitable, amount)
    // debit and credit actions also implement "CurrencyAmount" which means that they can be passed as arguments
    // they always return 0 "Amount()" until they are performed
    
    // note how debit action is passed as amount
    credit, _ := transfer.NewCreditAction(creditable, debit) 
    
    // debit and credit actions ignore to perform actual debit/credit in case if they receive 0 amount
    // which means that if debit is not performed then credit will not be performed as well
    // it is very important feature

    // order is important in this regard
    debit.Perform()
    credit.Perform()
```

## What about fees?

A common task is the fee, which should be included in the debit amount - when we need to subtract the fee from the amount that
we debited from the account. We must transfer part of this amount to one wallet, and the other part to another.
The package provides the "AmountConsumable" type for this purpose. 
This type allows you to subtract from the amount until its balance is greater than zero.


```go
    // source wallet with 100 euro
    sourceWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.NewFromInt(100)), *euro)
    // empty destination euro wallet
    destinationWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *euro)
    // empty revenue euro wallet
    revenueWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *euro)
    // debit amount is 30 euro
    debitAmount := transfers.NewAmount(*euro, decimal.NewFromInt(30))
    
    // take 30 euro from source wallet
    debit, _ := transfer.NewDebitAction(sourceWallet, debitAmount)
    // wrap the result as consumable amount
    debitReminder := transfer.NewAmountConsumable(debit)
    // consume 10% of the debit amount as fee
    fee, _ := transfer.NewDebitAction(debitReminder, transfer.NewAmountMultiplier(debitReminder, tenPercent))
    // credit the reminder (30 - 3) to the destination wallet
    credit, _ := transfer.NewCreditAction(destinationWallet, debitReminder)
    // credit fee to the revenue wallet
    creditFee, _ := transfer.NewCreditAction(revenueWallet, fee)

    // perform in the same order
	for _, action := range []Action{debit, fee, credit, creditFee} {
		err := action.Perform()
		if err != nil {
            return err
        }   
	}
```

See also `internal/transfer/fee/transfer.go`. There is implementation of transfer fee.

## Transfer in different currencies

So, what we have in the previous example already looks like solving a real problem. The next necessary element is processing
transfers with different currencies. As mentioned, the package provides "ExchangeAction" to solve this problem.
Since the "Action" can act as a "CurrencyAmount", we can calculate the value in another currency, and then simply pass
this action as an argument. In order to get the currency rate additional package "exchange" is used, its type "RateSource".
```go
// RateSource
type RateSource interface {
	FindRate(base, reference string) (Rate, error)
}
```

Below is example of transfer from "euroWallet" to "usdWallet".
```go
    euro := transfer.NewCurrency("EUR", 2)
    usd := transfer.NewCurrency("USD", 2)
    // debit amount is 10 EUR
    debitAmountEur := transfer.NewAmount(*euro, decimal.NewFromInt(10))

    // wallet with 100 euro
    euroWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.NewFromInt(100)), *euro)
    // empty usd wallet
    usdWallet :=  transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *usd)
    
    // debit 10 euro from euroWallet
    debit, _ := transfer.NewDebitAction(euroWallet, debitAmountEur)
    // now it is required to exchange debit amount in euro to amount in usd in order to be able to credit usdWallet
    // "exchange" package provides all necessary for that   
    // just for example let's assume that rate is constant value EUR/USD = 0.9
    rateEurUsd := exchange.NewRate("EUR", "USD", decimal.NewFromFloat(0.9))
    rateSource := exchange.NewDirectRateSource()
    rateSource.Set(rateEurUsd)
    
    // now as we have rate source we can proceed with exchange action
    // exchange debit(amount) using rateSource to usd
    exchange := transfer.NewExchangeAction(debit, rateSource, usd)
    
    // credit "exchange" amount in usd to usdWallet
    credit, _ := transfer.NewCreditAction(usdWallet, exchange)
    
    // now perform all actions
	for _, action := range []Action{debit, exchange, credit} {
		err := action.Perform()
		if err != nil {
            return err
        }   
	}

    // By the way. Package provides helper "performerGroup" which allows to do the same as above but with less code:
    err := transfer.NewPerformerGroup(debit, exchange, credit).Perform()
```

## Useful types

Transfer package provides several useful types which helps to organize the code and implement features.

**JoinCreditable** and **JoinDebitable** help to join creditable/debitable so that they act as a single instance.
It may be useful when you need to deal with different balances that must be updated in the same time. For example current and
available balance. 
```go
    //...
    bothBalances, _ := transfer.JoinDebitable(availableBalance, currentBalance) 
    //...    
    debit, _ := transfer.NewDebitAction(bothBalances, debitAmount)
```
The same applies to creditable.


**HookAction** helps to intercept the action before being executed.
```go
	debit, _ := transfer.NewDebitAction(sourceWallet, amount)	
	debitWithHook := transfer.NewHookAction(debit, func(action transfer.Action) error {
        var err error
		// decide whether to perform action
        if someCondition {
            err = action.Perform()
        }
        // do something else
        return err
	})    
```

**NewRoundAmount** function accepts "CurrencyAmount" and rounding function. It produces 2 "CurrencyAmount"s first is
rounded value, and the second is reminder (difference between original value, and the rounded value).

**LinkedBalance** the same as "Balance" but accepts initial balance value by reference. It is useful when you need to bind
some value in order to check the result later.

```go
    type Account struct {
        Number string
        Balance decimal.Decimal
    }

    myAccount := &Account{Number: "12345", Balance: decimal.NewFromInt(100)}
    euro := transfer.NewCurrency("EUR", 2)

    balance := transfer.NewLinkedBalance(&myAccount.Balance)
    wallet := transfer.NewWallet(balance, *euro)

    debit, _ := transfer.NewDebitAction(wallet, transfer.NewAmount(euro, decimal.NewFromInt(10)))
    debit.Perform()
    
    fmt.Println(myAcction.Balance) // 90
    
```

## Builder

Although the examples described are simple, building the required sequence of actions can be very verbose and not convenient.
To simplify this task, the "transfer" package has a "builder" subpackage.
If you've ever worked with SQL builder, the concept is already familiar to you. Builder is a framework that provides a simple
interface for creating a queue of actions. Under the hood, this package uses all the same elements we already talked about.

Here is an example of how you can build a simple transfer using the "builder" package.


```go
	euro := transfer.NewCurrency("EUR", 2)
	// wallet with 100 euro
	sourceWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.NewFromInt(100)), *euro)
    // empty wallet
	destinationWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *euro)

    // debit 70 euro from the source wallet, remember it as "debit", credit the amount remembered as "debit" to the destination wallet
	myTransfer := builder.
		Debit(70).
		From(sourceWallet).
		As("debit").
		CreditFromAlias("debit").
		To(destinationWallet)

	err := myTransfer.Execute()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sourceWallet.Amount()) // 30
	fmt.Println(destinationWallet.Amount()) // 70
```


Builder is very useful when it comes to handle conditional actions. 
For example, we may apply fee or not based on some condition.


```go
	euro := transfer.NewCurrency("EUR", 2)
	// wallet with 100 euro
	sourceWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.NewFromInt(100)), *euro)
	destinationWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *euro)
	feeWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *euro)

	myTransfer := builder.
		Debit(70).
		From(sourceWallet).
		As("debit")
    
    // apply fee if current timestamp is even
	if time.Now().Nanosecond() % 2 == 0 {
		myTransfer.
			Debit(10).
			From(sourceWallet).
			As("even timestamp fee").
			CreditFromAlias("even timestamp fee").
			To(feeWallet)
	}

	myTransfer.
		CreditFromAlias("debit").
		To(destinationWallet)

	err := myTransfer.Execute()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("sourceWallet.Amount() = %s\n", sourceWallet.Amount()) // 20 or 30
	fmt.Printf("destinationWallet.Amount() = %s\n", destinationWallet.Amount()) // 70
	fmt.Printf("feeWallet.Amount() = %s\n", feeWallet.Amount()) // 0 or 10
```

One of the useful features is grouping actions.


```go
	euro := transfer.NewCurrency("EUR", 2)
	// wallet with 100 euro
	sourceWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.NewFromInt(100)), *euro)
	destinationWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *euro)

	myTransfer := builder.
		Debit(70).
		From(sourceWallet).
		As("debit").
		IncludeToGroup("totalDebit")

	myTransfer.
		Debit(10).
		From(sourceWallet).
		As("debit2").
		IncludeToGroup("totalDebit").
		CreditFromAlias("debit2").
		To(destinationWallet)


	myTransfer.Execute()

	fmt.Println(myTransfer.GetGroup("totalDebit").Sum()) // -80

	// prints:
	// Currency: EUR, Amount: -70
	// Currency: EUR, Amount: -10
	for _, action := range myTransfer.GetGroup("totalDebit") {
		currency := action.Currency()
		sign := action.Sign()
		amount := action.Amount()

		fmt.Printf("Currency: %s, Amount: %s\n", currency.String(), amount.Mul(decimal.NewFromInt(int64(sign))) )
	}
```

Example of building transfer with exchange action.


```go
	euro := transfer.NewCurrency("EUR", 2)
	usd := transfer.NewCurrency("USD", 2)
	// wallet with 100 euro
	sourceWalletEur := transfer.NewWallet(transfer.NewSimpleBalance(decimal.NewFromInt(100)), *euro)
	destinationWalletUsd := transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *usd)

	rateEurUsd := exchange.NewRate("EUR", "USD", decimal.NewFromFloat(1.1))
	rateSource := exchange.NewDirectRateSource()
	rateSource.Set(rateEurUsd)

	myTransfer := builder.
		Debit(10).
		From(sourceWalletEur).
		As("debitInEur").
		ExchangeFromAlias("debitInEur").
		Using(rateSource).
		ToCurrency(usd).
		As("debitInUsd").
		CreditFromAlias("debitInUsd").
		To(destinationWalletUsd)

	myTransfer.Execute()

	fmt.Printf("sourceWalletEur.Amount() = %s\n", sourceWalletEur.Amount()) // 90
	fmt.Printf("destinationWalletUsd.Amount() = %s\n", destinationWalletUsd.Amount()) // 11
```

Builder also has the ability to add a callback (HookAction) when building a transfer.


```go
    /* Output:
        2020/10/30 18:03:28 debit action is performed
        2020/10/30 18:03:28 credit action is performed
        30
        70
   */
	euro := transfer.NewCurrency("EUR", 2)
	// wallet with 100 euro
	sourceWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.NewFromInt(100)), *euro)
	// empty wallet
	destinationWallet := transfer.NewWallet(transfer.NewSimpleBalance(decimal.Zero), *euro)

	// debit 70 euro from the source wallet, remember it as "debit", credit the amount remembered as "debit" to the destination wallet
	myTransfer := builder.
		Debit(70).
		From(sourceWallet).
		As("debit").
		WithCallback(func(action transfer.Action) error {
			err := action.Perform()
			log.Println("debit action is performed")
			return err
		}).
		CreditFromAlias("debit").
		To(destinationWallet).
		WithCallback(func(action transfer.Action) error {
			err := action.Perform()
			log.Println("credit action is performed")
			return err
		})

	err := myTransfer.Execute()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sourceWallet.Amount())      // 30
	fmt.Println(destinationWallet.Amount()) // 70
```