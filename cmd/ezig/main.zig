const std = @import("std");
const os = std.os;
const fs = std.fs;
const mem = std.mem;
const net = std.net;

const usage =
    \\Usage: ezig [command] [options]
    \\
    \\Commands:
    \\
    \\  help      print this message and exit
    \\  list      list available files
    \\  download  download file
    \\  tracker   set/get tracker address
    \\
    \\General options:
    \\
    \\  -h, --help      Print command-specific usage
    \\
;

pub fn main() !void {
    var gpa_instance = std.heap.GeneralPurposeAllocator(.{}){};
    defer {
        _ = gpa_instance.deinit();
    }
    const gpa = gpa_instance.allocator();

    var arena_instance = std.heap.ArenaAllocator.init(gpa);
    defer arena_instance.deinit();
    const arena = arena_instance.allocator();

    const args = try std.process.argsAlloc(arena);

    if (args.len <= 1) {
        std.log.info("{s}", .{usage});
        fatal("expected command argument", .{});
    }

    const cmd = findCommand(args[1]) orelse args[1];
    const cmd_args = args[2..];

    if (mem.eql(u8, cmd, "help") or
        mem.eql(u8, cmd, "--help") or
        mem.eql(u8, cmd, "-h"))
    {
        return std.io.getStdOut().writeAll(usage);
    } else if (mem.eql(u8, cmd, "list")) {
        return cmdList(arena);
    } else if (mem.eql(u8, cmd, "download")) {
        return cmdDownload(cmd_args);
    } else if (mem.eql(u8, cmd, "tracker")) {
        return cmdTracker(arena, cmd_args);
    } else {
        fatal("unknown command: {s}", .{cmd});
    }
}

// each cmd should start with a unique letter
const commands = [_][]const u8{
    "help",
    "list",
    "download",
    "tracker",
};

fn findCommand(needle: []const u8) ?[]const u8 {
    for (commands) |c| {
        if (mem.startsWith(u8, c, needle)) {
            return c;
        }
    }
    return null;
}

fn cmdList(all: mem.Allocator) !void {
    const tracker_addr = getTrackerAddr(all) catch |err| {
        if (err == error.FileNotFound) {
            fatal("tracker not set", .{});
        } else {
            fatal("unexpected error: {}", .{err});
        }
    };
    defer all.free(tracker_addr);

    const conn = net.tcpConnectToHost(all, tracker_addr, 22200) catch |err| {
        fatal("{}", .{err});
    };
    defer conn.close();

    const req =
        "GET /?id=all HTTP/1.1\r\n" ++
        "Host: localhost:22200\r\n" ++
        "User-Agent: ezig\r\n" ++
        "Accept: */*\r\n" ++
        "\r\n";

    const nwritten = try conn.write(req);
    std.log.info("wrote: {d}", .{nwritten});

    var buf: [8196]u8 = undefined;
    const nread = try conn.read(&buf);
    std.log.info("read: {d}", .{nread});

    try std.io.getStdOut().writer().print("{s}\n", .{buf});
}

fn cmdDownload(args: []const []const u8) !void {
    _ = args;

    std.log.info("download", .{});
}

fn cmdTracker(all: mem.Allocator, args: []const []const u8) !void {
    if (args.len < 1) {
        const tracker_addr = getTrackerAddr(all) catch |err| {
            if (err == error.FileNotFound) {
                fatal("tracker not set", .{});
            } else {
                fatal("unexpected error: {}", .{err});
            }
        };
        defer all.free(tracker_addr);
        try std.io.getStdOut().writer().print("{s}\n", .{tracker_addr});
    } else {
        try setTrackerAddr(all, args[0]);
    }
}

const Error = error{
    HomeNotFound,
};

fn readFile(all: mem.Allocator, path: []const u8) ![]const u8 {
    const f = try std.fs.openFileAbsolute(path, std.fs.File.OpenFlags{});
    defer f.close();
    const st = try f.stat();
    const content = try f.reader().readAllAlloc(all, st.size);
    return content;
}

fn writeFile(path: []const u8, data: []const u8) !void {
    const f = try fs.createFileAbsolute(path, std.fs.File.CreateFlags{});
    defer f.close();
    try f.writer().writeAll(data);
}

// returned value must be freed by caller
fn createTrackerPath(all: mem.Allocator) ![]const u8 {
    const home_dir = os.getenv("HOME") orelse return error.HomeNotFound;
    return try fs.path.join(all, &[_][]const u8{ home_dir, ".ez" });
}

// returned value must be freed by caller
fn getTrackerAddr(all: mem.Allocator) ![]const u8 {
    const tracker_path = try createTrackerPath(all);
    defer all.free(tracker_path);
    return try readFile(all, tracker_path);
}

fn setTrackerAddr(all: mem.Allocator, value: []const u8) !void {
    const tracker_path = try createTrackerPath(all);
    defer all.free(tracker_path);
    try writeFile(tracker_path, value);
}

fn fatal(comptime format: []const u8, args: anytype) noreturn {
    std.log.err(format, args);
    std.process.exit(1);
}
