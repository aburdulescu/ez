const std = @import("std");

pub const Response = struct {
    status: []const u8, // e.g. "200 OK"
    status_code: u8, // e.g. 200
    proto: []const u8, // e.g. "HTTP/1.0"
    proto_major: u8, // e.g. 1
    proto_minor: u8, // e.g. 0

    headers_k: []const []const u8,
    headers_v: []const []const u8,

    body: []const u8,
};

const Query = struct {
    k: []const u8,
    v: []const u8,
};

pub fn get(allocator: std.mem.Allocator, host: []const u8, port: u16, path: []const u8, query: ?[]const Query) !Response {
    _ = allocator;
    _ = query;

    std.log.warn("http://{s}:{d}{s}", .{ host, port, path });

    return Response{
        .status = "200 OK",
        .status_code = 200,
        .proto = "HTTP/1.1",
        .proto_major = 1,
        .proto_minor = 1,
        .headers_k = undefined,
        .headers_v = undefined,
        .body = "foo",
    };
}

const testing = std.testing;

test "basic" {
    _ = try get(testing.allocator, "localhost", 22200, "/", null);
}
