# prometheus-http-sd
a RAFT based prometheus http sd provider

# API

```
GET    /api/v1/target/                                        # return targets list
POST   /api/v1/target/                                        # creates a new target group
GET    /api/v1/target/<target_group_id>                       # retrieves the target group
PUT    /api/v1/target/<target_group_id>/label                 # add a label or server to target group
PATCH  /api/v1/target/<target_group_id>/label/<label_key>     # updates a label in a target group
DELETE /api/v1/target/<target_group_id>/label/<label_key>     # deletes a label in a target group
DELETE /api/v1/target/<target_group_id>/server/<server_addr>  # deletes a server in a target group
```