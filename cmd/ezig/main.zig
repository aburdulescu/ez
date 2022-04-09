const std = @import("std");
const os = std.os;
const fs = std.fs;

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

    if (std.mem.eql(u8, cmd, "help") or
        std.mem.eql(u8, cmd, "--help") or
        std.mem.eql(u8, cmd, "-h"))
    {
        return std.io.getStdOut().writeAll(usage);
    } else if (std.mem.eql(u8, cmd, "list")) {
        return cmdList();
    } else if (std.mem.eql(u8, cmd, "download")) {
        return cmdDownload(cmd_args);
    } else if (std.mem.eql(u8, cmd, "tracker")) {
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
        if (std.mem.startsWith(u8, c, needle)) {
            return c;
        }
    }
    return null;
}

fn cmdList() !void {
    std.log.warn("list", .{});

    // get tracker addr
    // send req to tracker
    // print results given by tracker
}

fn cmdDownload(args: []const []const u8) !void {
    _ = args;

    std.log.warn("download", .{});
}

fn cmdTracker(all: std.mem.Allocator, args: []const []const u8) !void {
    if (args.len < 1) {
        const tracker_addr = try getTrackerAddr(all);
        defer all.free(tracker_addr);
        try std.io.getStdOut().writer().print("{s}\n", .{tracker_addr});
    } else {
        // test to see if valid addr(ip/hostname)?
        // set given tracker addr in ~/.ezig/tracker
    }
}

const Error = error{
    HomeNotFound,
};

fn readFile(all: std.mem.Allocator, path: []const u8) ![]const u8 {
    const f = try std.fs.openFileAbsolute(path, std.fs.File.OpenFlags{});
    defer f.close();
    const st = try f.stat();
    const content = try f.reader().readAllAlloc(all, st.size);
    return content;
}

fn getTrackerAddr(all: std.mem.Allocator) ![]const u8 {
    const home_dir = os.getenv("HOME") orelse return error.HomeNotFound;
    const tracker_path = try fs.path.join(all, &[_][]const u8{ home_dir, ".ez" });
    defer all.free(tracker_path);
    return try readFile(all, tracker_path);
}

fn fatal(comptime format: []const u8, args: anytype) noreturn {
    std.log.err(format, args);
    std.process.exit(1);
}
