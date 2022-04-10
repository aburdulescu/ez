const std = @import("std");

pub const Response = struct {
    status: []const u8, // e.g. "200 OK"
    status_code: int, // e.g. 200
    proto: []const u8, // e.g. "HTTP/1.0"
    proto_major: int, // e.g. 1
    proto_minor: int, // e.g. 0

    headers_k: []const []const u8,
    headers_v: []const []const u8,

    body: []const u8,
};

pub fn get(url: []const u8) !Response {
    unreachable;
}
