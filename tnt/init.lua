box.cfg{
    listen = 3301,
    work_dir = "tmp",
}

box.schema.user.create("operator", {
    if_not_exists = true,
    password = "operator_pass"
})

box.schema.space.create('keyvalue', {
    if_not_exists = true,
    format = {
        { name = 'key', type = 'string', nullable = false },
        { name = 'value', type = 'string', nullable = false },
        { name = 'timestamp', type = 'unsigned', nullable = false},
        { name = 'meta', type = 'string', nullable = true },
    }
})

box.space.keyvalue:create_index("primary", {
    parts = {"key"},
    unique = true,
    if_not_exists = true,
})

box.schema.user.grant("operator", "read,write,execute", "universe")
