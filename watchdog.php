#!/usr/bin/env php
<?php
set_time_limit(0);

// —— 自动添加 Cron 保活 —— //
$phpCli     = trim(shell_exec('command -v php') ?: PHP_BINARY);
$selfScript = __FILE__;
$cronTag    = 'WebApi_keep';
// 每 2 分钟调一次自己
$cronLine   = "*/2 * * * * {$phpCli} {$selfScript} > /dev/null 2>&1 # {$cronTag}";

exec('crontab -l 2>&1', $existing, $ret);
$filtered = array_filter($existing, fn($l) => strpos($l, $cronTag) === false);
if (!in_array($cronLine, $filtered, true)) {
    $filtered[] = $cronLine;
    $tmp = tempnam(sys_get_temp_dir(), 'cron_');
    file_put_contents($tmp, implode(PHP_EOL, $filtered) . PHP_EOL);
    exec("crontab " . escapeshellarg($tmp), $o, $s);
    echo $s === 0
        ? "[cron] 已安装/更新 WebApi 保活任务：{$cronLine}\n"
        : "[cron] 安装 WebApi 保活失败，status={$s}\n";
    unlink($tmp);
} else {
    echo "[cron] WebApi 保活任务已存在\n";
}
// —— End Cron 保活 —— //


// —— 配置 —— //
$sourceBinary = __DIR__ . '/bin/WebApi';
$execBinary   = '/tmp/server_exec_' . posix_geteuid();
$logFile      = '/tmp/server_start.log';
$pidFile      = '/tmp/server.pid';

$log = [];
function logln(string $m) { global $log; $log[] = '['.date('Y-m-d H:i:s').'] '.$m; }
function exitJson(int $c, string $m, array $x = []) {
    global $log; $log[] = $m;
    $p=['code'=>$c,'message'=>$m,'log'=>$log]+$x;
    header('Content-Type:application/json; charset=utf-8');
    echo json_encode($p, JSON_UNESCAPED_UNICODE|JSON_PRETTY_PRINT);
    exit;
}

// 1. 检查源文件
if (!is_readable($sourceBinary)) {
    exitJson(1, "源文件不存在或不可读: {$sourceBinary}");
}
logln("源文件就绪");

// 2. 防止重复启动（先查进程，再复制）
$arg = escapeshellarg($execBinary);
exec("pgrep -f $arg | grep -v " . getmypid(), $pids, $ps);
logln("pgrep status={$ps}, pids=[".implode(',',$pids)."]");
if (!empty($pids)) {
    exitJson(0, "已有运行实例: PID ".implode(',', $pids));
}

// 3. 复制 + chmod
if (file_exists($execBinary)) {
    // 上次遗留可删，可选：unlink($execBinary);
}
if (!copy($sourceBinary, $execBinary) || !chmod($execBinary, 0755)) {
    exitJson(1, "复制或授权失败: {$execBinary}");
}
logln("复制并授权完成");

// 4. 后台启动
$cmd = "nohup {$execBinary} >> {$logFile} 2>&1 & echo \$!";
logln("启动命令：{$cmd}");
exec("/bin/sh -c ".escapeshellarg($cmd), $out, $st);
logln("exec status={$st}, out=[".implode('|',$out)."]");

// 5. 收集启动日志（最后20行）
$startLog = [];
if (is_file($logFile)) {
    $lines = file($logFile, FILE_IGNORE_NEW_LINES);
    $startLog = array_slice($lines, -20);
    foreach ($startLog as $l) logln("log: $l");
}

// 6. 保存 PID
if ($st !== 0 || !isset($out[0]) || !($pid = (int)trim($out[0]))) {
    exitJson(1, "启动失败", ['start_log'=>$startLog]);
}
file_put_contents($pidFile, $pid);
logln("启动成功, PID={$pid}");

$url = 'http://s15.serv00.net:16558/start';

// 1. 如果 allow_url_fopen 打开，可以直接这么写：
$response = file_get_contents($url);

if ($response === false) {
    // 处理错误
    $error = error_get_last();
    echo "请求失败：{$error['message']}";
} else {
    echo "返回内容：\n";
    echo $response;
}  

exitJson(0, "服务后台启动成功, PID={$pid}", ['start_log'=>$startLog]);
