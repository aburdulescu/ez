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

const Query = struct {
    k: []const u8,
    v: []const u8,
};

pub fn get(allocator: std.mem.Allocator, host: []const u8, port: u16, path: []const u8, query: []const Query) !Response {
    unreachable;
}
