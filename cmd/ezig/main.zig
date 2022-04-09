const std = @import("std");

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
        return cmdList(cmd_args);
    } else if (std.mem.eql(u8, cmd, "download")) {
        return cmdDownload(cmd_args);
    } else if (std.mem.eql(u8, cmd, "tracker")) {
        return cmdTracker(cmd_args);
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

fn cmdList(args: []const []const u8) !void {
    _ = args;
}

fn cmdDownload(args: []const []const u8) !void {
    _ = args;
}

fn cmdTracker(args: []const []const u8) !void {
    _ = args;
}

fn fatal(comptime format: []const u8, args: anytype) noreturn {
    std.log.err(format, args);
    std.process.exit(1);
}
