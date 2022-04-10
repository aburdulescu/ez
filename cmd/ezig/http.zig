const std = @import("std");
const log = std.log;
const net = std.net;

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

fn buildQueryString(queries: []const Query) ![]const u8 {}

const user_agent = "ezhttp";

pub fn get(allocator: std.mem.Allocator, host: []const u8, port: u16, path: []const u8, query: ?[]const Query) !Response {
    _ = allocator;
    _ = query;

    const conn = try net.tcpConnectToHost(allocator, host, 22200);
    defer conn.close();

    const req_fmt =
        "GET {s}{s} HTTP/1.1\r\n" ++
        "Host: {s}:{d}\r\n" ++
        "User-Agent: " ++ user_agent ++ "\r\n" ++
        "Accept: */*\r\n" ++
        "\r\n";

    // build query

    try std.fmt.format(conn.writer(), req_fmt, .{ path, "", host, port });

    var buf: [8196]u8 = undefined;
    const nread = try conn.read(&buf);
    std.log.info("read: {d}", .{nread});

    try std.io.getStdOut().writer().print("{s}\n", .{buf});

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

test "without query" {
    _ = try get(testing.allocator, "localhost", 22200, "/", null);
}

test "with query" {
    const query = &[_]Query{
        .{ .k = "id", .v = "all" },
    };
    _ = try get(testing.allocator, "localhost", 22200, "/", query);
}
