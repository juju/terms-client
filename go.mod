module github.com/juju/terms-client/v2

go 1.16

require (
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/juju/ansiterm v0.0.0-20180109212912-720a0952cc2a // indirect
	github.com/juju/charm/v8 v8.0.0-20200925053015-07d39c0154ac
	github.com/juju/cmd v0.0.0-20200108104440-8e43f3faa5c9
	github.com/juju/errors v0.0.0-20200330140219-3fe23663418f
	github.com/juju/gnuflag v0.0.0-20171113085948-2ce1bb71843d
	github.com/juju/idmclient/v2 v2.0.0-20210305130631-95f6b39e9f5d
	github.com/juju/juju v0.0.0-20201007080928-1f35f6a20b57
	github.com/juju/persistent-cookiejar v0.0.0-20170428161559-d67418f14c93
	github.com/juju/testing v0.0.0-20200923013621-75df6121fbb0
	github.com/juju/usso v1.0.1 // indirect
	github.com/juju/utils/v2 v2.0.0-20200923005554-4646bfea2ef1
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/juju/environschema.v1 v1.0.0
	gopkg.in/macaroon-bakery.v3 v3.0.0-20210305063614-792624751518
	gopkg.in/retry.v1 v1.0.3 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

// Needed for github.com/juju/juju
replace github.com/altoros/gosigma => github.com/juju/gosigma v0.0.0-20200420012028-063911838a9e

replace gopkg.in/mgo.v2 => github.com/juju/mgo v2.0.0-20190418114320-e9d4866cb7fc+incompatible

replace github.com/hashicorp/raft => github.com/juju/raft v2.0.0-20200420012049-88ad3b3f0a54+incompatible
