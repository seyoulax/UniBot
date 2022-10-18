<?php 
header("Access-Control-Allow-Origin: *");
$host = 'mysql';
$db   = 'bot_data';
$user = 'root';
$pass = 'inordic123';
$charset = 'utf8';
$dsn = "mysql:host=$host;dbname=$db;charset=$charset";
$opt = [
    PDO::ATTR_ERRMODE            => PDO::ERRMODE_EXCEPTION,
    PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
    PDO::ATTR_EMULATE_PREPARES   => false,
];
$pdo = new PDO($dsn, $user, $pass, $opt);
$data = $pdo->query("SELECT `id`, `first_name`, `username`, `latitude`, `longitude` FROM `users`");
$bigData = [];
while($row = $data->fetch()){
    $bigData[] = $row;
};
echo json_encode($bigData);
