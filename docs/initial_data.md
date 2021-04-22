**Initial data**

The DB has default filled data.  

***Balance types***  

|Id | name              |
|---|-------------------|
| 1 | account           |
| 2 | card              |
| 3 | revenue_account   |

***Card type categories***  

|Id | name              |
|---|-------------------|
| 1 | Manual            |
| 2 | Integrated        |
 
***Card type formats***  
 
|Id | name              | code              |
|---|-------------------|-------------------|
| 1 |Manual             |alphanumeric       |
| 2 |Integrated         |sixteen_numeric    |

***Payment periods***  

|Id | name              |
|---|-------------------|
| 1 | Monthly           |
| 2 | Quarterly         |
| 3 | Bi-Annually       |
| 4 | Annually          |

***Payout methods***  
 
|Id | name                      | method            |
|---|---------------------------|-------------------|
| 1 |Daily interest calculation |daily              |

***Request subjects***  
 
| subject | description       |
|---------|-------------------|
|CA |Credit Account           |
|CFT|Card Funding Transfer    |
|DA |Debit Account            |
|DRA|Deduct Revenue Account   |
|IWT|Incoming Wire Transfer   |
|OWT|Outgoing Wire Transfer   |
|TBA|Transfer Between Accounts|
|TBU|Transfer Between Users   |

***Settings***  
 
|Id | name                      | value             | description       |
|---|---------------------------|-------------------|-------------------|
| 1 |tba_action_required |true|Transfer between accounts. Administrator will receive notification about transfer and will have to execute/cancel it.              |
| 2 |tba_tan_required |false |Transfer between accounts. User will have to enter TAN for transfer to proceed.|
| 3 |tbu_action_required |true|Transfer between users. Administrator will receive notification about transfer and will have to execute/cancel it.|
| 4 |tbu_tan_required |false|Transfer between users. User will have to enter TAN for transfer to proceed.|
| 5 |owt_action_required |true|Outgoing wire transfer. Administrator will receive notification about transfer and will have to execute/cancel it.|
| 6 |owt_tan_required |false|Outgoing wire transfer. User will have to enter TAN for transfer to proceed.|
| 7 |cft_action_required |true|Card funding transfer. Administrator will receive notification about transfer and will have to execute/cancel it.|
| 8 |cft_tan_required |false|Card funding transfer. User will have to enter TAN for transfer to proceed.|
| 9 |tan_generate_qty |10|Number of tans to generate. Standard TANs Qty.|
| 10 |tan_generate_trigger_qty |5|Number of remaining tans to trigger generation. TANs Qty Remaining Limit.|
| 11 |tan_message_subject |TANs              |Message subject              |
| 12 |tan_message_content |Please, copy or print this message, since it is only going to be shown once. Your TANs: [Tan]              |              |

**Countries**  

|Id |Name                                        |code |code3|code_numeric|
|---|--------------------------------------------|---|---|---|
|1  |Afghanistan                                 |AF |AFG|004|
|2  |Aland Islands                               |AX |ALA|248|
|3  |Albania                                     |AL |ALB|008|
|4  |Algeria                                     |DZ |DZA|012|
|5  |American Samoa                              |AS |ASM|016|
|6  |Andorra                                     |AD |AND|020|
|7  |Angola                                      |AO |AGO|024|
|8  |Anguilla                                    |AI |AIA|660|
|9  |Antarctica                                  |AQ |ATA|010|
|10 |Antigua and Barbuda                         |AG |ATG|028|
|11 |Argentina                                   |AR |ARG|032|
|12 |Armenia                                     |AM |ARM|051|
|13 |Aruba                                       |AW |ABW|533|
|14 |Australia                                   |AU |AUS|036|
|15 |Austria                                     |AT |AUT|040|
|16 |Azerbaijan                                  |AZ |AZE|031|
|17 |Bahamas                                     |BS |BHS|044|
|18 |Bahrain                                     |BH |BHR|048|
|19 |Bangladesh                                  |BD |BGD|050|
|20 |Barbados                                    |BB |BRB|052|
|21 |Belarus                                     |BY |BLR|112|
|22 |Belgium                                     |BE |BEL|056|
|23 |Belize                                      |BZ |BLZ|084|
|24 |Benin                                       |BJ |BEN|204|
|25 |Bermuda                                     |BM |BMU|060|
|26 |Bhutan                                      |BT |BTN|064|
|27 |Bolivia                                     |BO |BOL|068|
|28 |Bosnia and Herzegovina                      |BA |BIH|070|
|29 |Botswana                                    |BW |BWA|072|
|30 |Bouvet Island                               |BV |BVT|074|
|31 |Brazil                                      |BR |BRA|076|
|32 |British Virgin Islands                      |VG |VGB|092|
|33 |British Indian Ocean Territory              |IO |IOT|086|
|34 |Brunei Darussalam                           |BN |BRN|096|
|35 |Bulgaria                                    |BG |BGR|100|
|36 |Burkina Faso                                |BF |BFA|854|
|37 |Burundi                                     |BI |BDI|108|
|38 |Cambodia                                    |KH |KHM|116|
|39 |Cameroon                                    |CM |CMR|120|
|40 |Canada                                      |CA |CAN|124|
|41 |Cape Verde                                  |CV |CPV|132|
|42 |Cayman Islands                              |KY |CYM|136|
|43 |Central African Republic                    |CF |CAF|140|
|44 |Chad                                        |TD |TCD|148|
|45 |Chile                                       |CL |CHL|152|
|46 |China                                       |CN |CHN|156|
|47 |Hong Kong, SAR China                        |HK |HKG|344|
|48 |Macao, SAR China                            |MO |MAC|446|
|49 |Christmas Island                            |CX |CXR|162|
|50 |Cocos (Keeling) Islands                     |CC |CCK|166|
|51 |Colombia                                    |CO |COL|170|
|52 |Comoros                                     |KM |COM|174|
|53 |Congo (Brazzaville)                         |CG |COG|178|
|54 |Congo, (Kinshasa)                           |CD |COD|180|
|55 |Cook Islands                                |CK |COK|184|
|56 |Costa Rica                                  |CR |CRI|188|
|57 |Côte d'Ivoire                               |CI |CIV|384|
|58 |Croatia                                     |HR |HRV|191|
|59 |Cuba                                        |CU |CUB|192|
|60 |Cyprus                                      |CY |CYP|196|
|61 |Czech Republic                              |CZ |CZE|203|
|62 |Denmark                                     |DK |DNK|208|
|63 |Djibouti                                    |DJ |DJI|262|
|64 |Dominica                                    |DM |DMA|212|
|65 |Dominican Republic                          |DO |DOM|214|
|66 |Ecuador                                     |EC |ECU|218|
|67 |Egypt                                       |EG |EGY|818|
|68 |El Salvador                                 |SV |SLV|222|
|69 |Equatorial Guinea                           |GQ |GNQ|226|
|70 |Eritrea                                     |ER |ERI|232|
|71 |Estonia                                     |EE |EST|233|
|72 |Ethiopia                                    |ET |ETH|231|
|73 |Falkland Islands (Malvinas)                 |FK |FLK|238|
|74 |Faroe Islands                               |FO |FRO|234|
|75 |Fiji                                        |FJ |FJI|242|
|76 |Finland                                     |FI |FIN|246|
|77 |France                                      |FR |FRA|250|
|78 |French Guiana                               |GF |GUF|254|
|79 |French Polynesia                            |PF |PYF|258|
|80 |French Southern Territories                 |TF |ATF|260|
|81 |Gabon                                       |GA |GAB|266|
|82 |Gambia                                      |GM |GMB|270|
|83 |Georgia                                     |GE |GEO|268|
|84 |Germany                                     |DE |DEU|276|
|85 |Ghana                                       |GH |GHA|288|
|86 |Gibraltar                                   |GI |GIB|292|
|87 |Greece                                      |GR |GRC|300|
|88 |Greenland                                   |GL |GRL|304|
|89 |Grenada                                     |GD |GRD|308|
|90 |Guadeloupe                                  |GP |GLP|312|
|91 |Guam                                        |GU |GUM|316|
|92 |Guatemala                                   |GT |GTM|320|
|93 |Guernsey                                    |GG |GGY|831|
|94 |Guinea                                      |GN |GIN|324|
|95 |Guinea-Bissau                               |GW |GNB|624|
|96 |Guyana                                      |GY |GUY|328|
|97 |Haiti                                       |HT |HTI|332|
|98 |Heard and Mcdonald Islands                  |HM |HMD|334|
|99 |Holy See (Vatican City State)               |VA |VAT|336|
|100|Honduras                                    |HN |HND|340|
|101|Hungary                                     |HU |HUN|348|
|102|Iceland                                     |IS |ISL|352|
|103|India                                       |IN |IND|356|
|104|Indonesia                                   |ID |IDN|360|
|105|Iran, Islamic Republic of                   |IR |IRN|364|
|106|Iraq                                        |IQ |IRQ|368|
|107|Ireland                                     |IE |IRL|372|
|108|Isle of Man                                 |IM |IMN|833|
|109|Israel                                      |IL |ISR|376|
|110|Italy                                       |IT |ITA|380|
|111|Jamaica                                     |JM |JAM|388|
|112|Japan                                       |JP |JPN|392|
|113|Jersey                                      |JE |JEY|832|
|114|Jordan                                      |JO |JOR|400|
|115|Kazakhstan                                  |KZ |KAZ|398|
|116|Kenya                                       |KE |KEN|404|
|117|Kiribati                                    |KI |KIR|296|
|118|Korea (North)                               |KP |PRK|408|
|119|Korea (South)                               |KR |KOR|410|
|120|Kuwait                                      |KW |KWT|414|
|121|Kyrgyzstan                                  |KG |KGZ|417|
|122|Lao PDR                                     |LA |LAO|418|
|123|Latvia                                      |LV |LVA|428|
|124|Lebanon                                     |LB |LBN|422|
|125|Lesotho                                     |LS |LSO|426|
|126|Liberia                                     |LR |LBR|430|
|127|Libya                                       |LY |LBY|434|
|128|Liechtenstein                               |LI |LIE|438|
|129|Lithuania                                   |LT |LTU|440|
|130|Luxembourg                                  |LU |LUX|442|
|131|Macedonia, Republic of                      |MK |MKD|807|
|132|Madagascar                                  |MG |MDG|450|
|133|Malawi                                      |MW |MWI|454|
|134|Malaysia                                    |MY |MYS|458|
|135|Maldives                                    |MV |MDV|462|
|136|Mali                                        |ML |MLI|466|
|137|Malta                                       |MT |MLT|470|
|138|Marshall Islands                            |MH |MHL|584|
|139|Martinique                                  |MQ |MTQ|474|
|140|Mauritania                                  |MR |MRT|478|
|141|Mauritius                                   |MU |MUS|480|
|142|Mayotte                                     |YT |MYT|175|
|143|Mexico                                      |MX |MEX|484|
|144|Micronesia, Federated States of             |FM |FSM|583|
|145|Moldova                                     |MD |MDA|498|
|146|Monaco                                      |MC |MCO|492|
|147|Mongolia                                    |MN |MNG|496|
|148|Montenegro                                  |ME |MNE|499|
|149|Montserrat                                  |MS |MSR|500|
|150|Morocco                                     |MA |MAR|504|
|151|Mozambique                                  |MZ |MOZ|508|
|152|Myanmar                                     |MM |MMR|104|
|153|Namibia                                     |NA |NAM|516|
|154|Nauru                                       |NR |NRU|520|
|155|Nepal                                       |NP |NPL|524|
|156|Netherlands                                 |NL |NLD|528|
|157|Netherlands Antilles                        |AN |ANT|530|
|158|New Caledonia                               |NC |NCL|540|
|159|New Zealand                                 |NZ |NZL|554|
|160|Nicaragua                                   |NI |NIC|558|
|161|Niger                                       |NE |NER|562|
|162|Nigeria                                     |NG |NGA|566|
|163|Niue                                        |NU |NIU|570|
|164|Norfolk Island                              |NF |NFK|574|
|165|Northern Mariana Islands                    |MP |MNP|580|
|166|Norway                                      |NO |NOR|578|
|167|Oman                                        |OM |OMN|512|
|168|Pakistan                                    |PK |PAK|586|
|169|Palau                                       |PW |PLW|585|
|170|Palestinian Territory                       |PS |PSE|275|
|171|Panama                                      |PA |PAN|591|
|172|Papua New Guinea                            |PG |PNG|598|
|173|Paraguay                                    |PY |PRY|600|
|174|Peru                                        |PE |PER|604|
|175|Philippines                                 |PH |PHL|608|
|176|Pitcairn                                    |PN |PCN|612|
|177|Poland                                      |PL |POL|616|
|178|Portugal                                    |PT |PRT|620|
|179|Puerto Rico                                 |PR |PRI|630|
|180|Qatar                                       |QA |QAT|634|
|181|Réunion                                     |RE |REU|638|
|182|Romania                                     |RO |ROU|642|
|183|Russian Federation                          |RU |RUS|643|
|184|Rwanda                                      |RW |RWA|646|
|185|Saint-Barthélemy                            |BL |BLM|652|
|186|Saint Helena                                |SH |SHN|654|
|187|Saint Kitts and Nevis                       |KN |KNA|659|
|188|Saint Lucia                                 |LC |LCA|662|
|189|Saint-Martin (French part)                  |MF |MAF|663|
|190|Saint Pierre and Miquelon                   |PM |SPM|666|
|191|Saint Vincent and Grenadines                |VC |VCT|670|
|192|Samoa                                       |WS |WSM|882|
|193|San Marino                                  |SM |SMR|674|
|194|Sao Tome and Principe                       |ST |STP|678|
|195|Saudi Arabia                                |SA |SAU|682|
|196|Senegal                                     |SN |SEN|686|
|197|Serbia                                      |RS |SRB|688|
|198|Seychelles                                  |SC |SYC|690|
|199|Sierra Leone                                |SL |SLE|694|
|200|Singapore                                   |SG |SGP|702|
|201|Slovakia                                    |SK |SVK|703|
|202|Slovenia                                    |SI |SVN|705|
|203|Solomon Islands                             |SB |SLB|090|
|204|Somalia                                     |SO |SOM|706|
|205|South Africa                                |ZA |ZAF|710|
|206|South Georgia and the South Sandwich Islands|GS |SGS|239|
|207|South Sudan                                 |SS |SSD|728|
|208|Spain                                       |ES |ESP|724|
|209|Sri Lanka                                   |LK |LKA|144|
|210|Sudan                                       |SD |SDN|736|
|211|Suriname                                    |SR |SUR|740|
|212|Svalbard and Jan Mayen Islands              |SJ |SJM|744|
|213|Swaziland                                   |SZ |SWZ|748|
|214|Sweden                                      |SE |SWE|752|
|215|Switzerland                                 |CH |CHE|756|
|216|Syrian Arab Republic (Syria)                |SY |SYR|760|
|217|Taiwan, Republic of China                   |TW |TWN|158|
|218|Tajikistan                                  |TJ |TJK|762|
|219|Tanzania, United Republic of                |TZ |TZA|834|
|220|Thailand                                    |TH |THA|764|
|221|Timor-Leste                                 |TL |TLS|626|
|222|Togo                                        |TG |TGO|768|
|223|Tokelau                                     |TK |TKL|772|
|224|Tonga                                       |TO |TON|776|
|225|Trinidad and Tobago                         |TT |TTO|780|
|226|Tunisia                                     |TN |TUN|788|
|227|Turkey                                      |TR |TUR|792|
|228|Turkmenistan                                |TM |TKM|795|
|229|Turks and Caicos Islands                    |TC |TCA|796|
|230|Tuvalu                                      |TV |TUV|798|
|231|Uganda                                      |UG |UGA|800|
|232|Ukraine                                     |UA |UKR|804|
|233|United Arab Emirates                        |AE |ARE|784|
|234|United Kingdom                              |GB |GBR|826|
|235|United States of America                    |US |USA|840|
|236|US Minor Outlying Islands                   |UM |UMI|581|
|237|Uruguay                                     |UY |URY|858|
|238|Uzbekistan                                  |UZ |UZB|860|
|239|Vanuatu                                     |VU |VUT|548|
|240|Venezuela (Bolivarian Republic)             |VE |VEN|862|
|241|Viet Nam                                    |VN |VNM|704|
|242|Virgin Islands, US                          |VI |VIR|850|
|243|Wallis and Futuna Islands                   |WF |WLF|876|
|244|Western Sahara                              |EH |ESH|732|
|245|Yemen                                       |YE |YEM|887|
|246|Zambia                                      |ZM |ZMB|894|
|247|Zimbabwe                                    |ZW |ZWE|716|
|248|Curacao                                     |CW |CUW|531|
