# glimit
A Go Rate Limiter backed by gorm

# caveats
* It doesn't utilize a monotonic clock, so hopefully your system clock is relatively sane.
* I really need to make it use associations to keep the db from becoming a mess.
* It's missing a lot of safety checks (you could accidentally delete a lot of rate limiting data)
* There's not much keeping you from "losing" a limiter and accumulating garbage, for example if you associated it with user sessions, you may want to do some housekeeping yourself.
* It has not been tested with very high rates yet, I wouldn't try it on anything mission critical, for casual use it should be good enough.
* I need to add stores other than gorm

# docs
* The comments and godoc should be enough to get you started
