[![Build Status](https://travis-ci.org/blasphemy/glimit.svg?branch=master)](https://travis-ci.org/blasphemy/glimit)
[![Coverage Status](https://coveralls.io/repos/github/blasphemy/glimit/badge.svg?branch=master)](https://coveralls.io/github/blasphemy/glimit?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/blasphemy/glimit)](https://goreportcard.com/report/github.com/blasphemy/glimit)
# glimit
A Go Rate Limiter backed by gorm

# caveats
* It doesn't utilize a monotonic clock, so hopefully your system clock is relatively sane.
* There's not much keeping you from "losing" a limiter and accumulating garbage, for example if you associated it with user sessions, you may want to do some housekeeping yourself.
* It has not been tested with very high rates yet, I wouldn't try it on anything mission critical, for casual use it should be good enough.
* I need to add an interface to use stores other than gorm
* It does not supply a blocking function, as that could could cause problems. If you want one, I suggest you implement one yourself in your application with the required precision.

# docs
* The comments and godoc should be enough to get you started
