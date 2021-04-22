package service

import (
	"strings"

	modelIwt "github.com/Confialink/wallet-accounts/internal/modules/bank-details/model"

	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

type Pdf struct{}

func NewPdf() *Pdf {
	return &Pdf{}
}

func (p *Pdf) CreatePdfBin(account *model.Account, iwt *modelIwt.IwtBankAccountModel) ([]byte, error) {
	var intermediaryHtml string
	if iwt.IntermediaryBankDetails != nil {
		intermediaryHtml = p.getBankDetailsHtml("Intermediary Bank Details", iwt.IntermediaryBankDetails)
	}

	var beneficiaryHtml string
	if iwt.BeneficiaryBankDetails != nil {
		beneficiaryHtml = p.getBankDetailsHtml("Benificiary Bank Details", iwt.BeneficiaryBankDetails)
	}
	creditToHtml := p.getCreditToHtml(iwt.BeneficiaryCustomer, account)

	html := `<html><head>
<meta content="text/html; charset=UTF-8" http-equiv="content-type">
<style>
body {
    font-family: sans-serif;
}
table {
    border-collapse: collapse;
    margin: 25px 0;
    font-size: 0.9em;
    font-family: sans-serif;
    width: 100%;
}
table th,
table td {
    padding: 12px 15px;
}
table tbody tr {
    border-bottom: 1px solid #dddddd;
}

table tbody tr:nth-of-type(even) {
    background-color: #f3f3f3;
}
</style>
</head>
<body>
<h2>Incoming Wire Transfer<h2>
<h4>Please fund your account using the following bank instructions:</h4>
<hr>
` + intermediaryHtml + `
` + beneficiaryHtml + `
` + creditToHtml + `
</body></html>`

	g, err := wkhtml.NewPDFGenerator()
	if err != nil {
		return nil, err
	}

	page := wkhtml.NewPageReader(strings.NewReader(html))

	page.FooterRight.Set("[page]")
	page.FooterFontSize.Set(10)
	page.Zoom.Set(0.95)

	// Add to document
	g.AddPage(page)

	if err := g.Create(); err != nil {
		return nil, err
	}
	return g.Bytes(), nil
}

func (p *Pdf) getBankDetailsHtml(title string, data *modelIwt.BankDetailsModel) string {
	return `
<h4>` + title + `</h4>
<table>
	<tbody>
        <tr>
			<td colspan="1" rowspan="1">Swift / BIC</td>
			<td colspan="1" rowspan="1">` + data.SwiftCode + `</td>
		</tr>
        <tr>
			<td colspan="1" rowspan="1">Bank Name</td>
			<td colspan="1" rowspan="1">` + data.BankName + `</td>
		</tr>
        <tr>
			<td colspan="1" rowspan="1">Address</td>
			<td colspan="1" rowspan="1">` + data.Address + `</td>
		</tr>
        <tr>
			<td colspan="1" rowspan="1">Location</td>
			<td colspan="1" rowspan="1">` + data.Location + `</td>
		</tr>
        <tr>
			<td colspan="1" rowspan="1">Country</td>
			<td colspan="1" rowspan="1">` + *data.Country.Name + `</td>
		</tr>
        <tr>
			<td colspan="1" rowspan="1">ABA/RTN</td>
			<td colspan="1" rowspan="1">` + data.AbaNumber + `</td>
		</tr>
    </tbody>
</table>
`
}

func (p *Pdf) getCreditToHtml(data *modelIwt.BeneficiaryCustomerModel, account *model.Account) string {
	return `
<h4>For Credit To</h4>
<table>
	<tbody>
        <tr>
			<td colspan="1" rowspan="1">Account name</td>
			<td colspan="1" rowspan="1">` + data.AccountName + `</td>
		</tr>
        <tr>
			<td colspan="1" rowspan="1">Address</td>
			<td colspan="1" rowspan="1">` + data.Address + `</td>
		</tr>
        <tr>
			<td colspan="1" rowspan="1">Acc#/IBAN</td>
			<td colspan="1" rowspan="1">` + data.Iban + `</td>
		</tr>
        <tr>
			<td colspan="1" rowspan="1">Reference / Message</td>
			<td colspan="1" rowspan="1">` + account.Number + `</td>
		</tr>
    </tbody>
</table>
`
}
