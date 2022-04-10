const std = @import("std");

pub const Response = struct {
    status: []const u8, // e.g. "200 OK"
    status_code: int, // e.g. 200
    proto: []const u8, // e.g. "HTTP/1.0"
    proto_major: int, // e.g. 1
    proto_minor: int, // e.g. 0

    // TODO: use map as Go does?
    headers: []const []const u8,

    body: []const u8, // e.g. 1

    content_length: i64,

    transfer_encoding: []const []const u8,

    close: bool,

    // uncompressed: bool,
};

pub fn get(url: []const u8) !Response {
    unreachable;
}
