wrk.method="GET"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["User-Agent"] = "PostmanRuntime/7.32.3"
-- 记得修改这个，你在登录页面登录一下，然后复制一个过来这里
wrk.headers["Authorization"]="Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjEsInVzZXJBZ2VudCI6IiIsImV4cCI6MTcyNTM0NjU0MX0.A5zdj2YPU6C04hpYWI1ggrmYR4Fbm2-yEM6vakQ9aQg"