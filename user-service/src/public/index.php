<?php
require '../vendor/autoload.php';

use App\Database;
use Firebase\JWT\JWT;
use Firebase\JWT\Key;

// Mengatur Header agar Client tahu ini JSON
header("Access-Control-Allow-Origin: *");
header("Content-Type: application/json; charset=UTF-8");
header("Access-Control-Allow-Methods: POST, GET, PUT, OPTIONS");
header("Access-Control-Allow-Headers: Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With");

// Handle Preflight Request 
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    exit;
}

// Konfigurasi JWT 
$secret_key = "KunciRahasiaSuperPanjangDanAcak123!@";
$algo = 'HS256';

// Koneksi Database
$database = new Database();
$db = $database->getConnection();

// Helper Response
function jsonResponse($data, $status = 200) {
    http_response_code($status);
    echo json_encode($data);
    exit;
}

// Router Sederhana
$method = $_SERVER['REQUEST_METHOD'];
$path = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);

// ROUTE PUBLIC (TIDAK BUTUH TOKEN)
// REGISTER
if ($method === 'POST' && $path === '/register') {
    $data = json_decode(file_get_contents("php://input"));
    
    if(!isset($data->name) || !isset($data->email) || !isset($data->password)) {
        jsonResponse(['message' => 'Data tidak lengkap (name, email, password wajib)'], 400);
    }

    $role = isset($data->role) && in_array($data->role, ['admin', 'user']) ? $data->role : 'user';
    $password_hash = password_hash($data->password, PASSWORD_BCRYPT);

    try {
        $stmt = $db->prepare("INSERT INTO users (name, email, password, role) VALUES (?, ?, ?, ?)");
        $stmt->execute([$data->name, $data->email, $password_hash, $role]);
        jsonResponse(['message' => 'User berhasil dibuat', 'role' => $role], 201);
    } catch (Exception $e) {
        jsonResponse(['message' => 'Email sudah terdaftar'], 409);
    }
}

// // LOGIN 
// if ($method === 'POST' && $path === '/login') {
//     $data = json_decode(file_get_contents("php://input"));

//     $stmt = $db->prepare("SELECT id, name, email, password, role FROM users WHERE email = ?");
//     $stmt->execute([$data->email]);
//     $user = $stmt->fetch(PDO::FETCH_ASSOC);

//     if ($user && password_verify($data->password, $user['password'])) {
//         $payload = [
//             'iss' => 'user-service',
//             'iat' => time(),
//             'exp' => time() + (60*60*24), 
//             'data' => [
//                 'id' => $user['id'],
//                 'role' => $user['role']
//             ]
//         ];

//         $jwt = JWT::encode($payload, $secret_key, $algo);
//         jsonResponse([
//             'message' => 'Login sukses',
//             'token' => $jwt,
//              'id' => $user['id'],
//             'role' => $user['role']
//         ]);
//     } else {
//         jsonResponse(['message' => 'Email atau password salah'], 401);
//     }
// }
// LOGIN 
if ($method === 'POST' && $path === '/login') {
    $data = json_decode(file_get_contents("php://input"));

    $stmt = $db->prepare("SELECT id, name, email, password, role FROM users WHERE email = ?");
    $stmt->execute([$data->email]);
    $user = $stmt->fetch(PDO::FETCH_ASSOC);

    if ($user && password_verify($data->password, $user['password'])) {
        $payload = [
            'iss' => 'user-service',
            'iat' => time(),
            'exp' => time() + (60 * 60 * 24),
            'data' => [
                'id' => $user['id'],
                'role' => $user['role']
            ]
        ];

        $jwt = JWT::encode($payload, $secret_key, $algo);

       
        jsonResponse([
            'message' => 'Login sukses',
            'token' => $jwt,
            'role' => $user['role'],
            'user' => [
                'id' => $user['id'],
                'name' => $user['name'],
                'email' => $user['email']
            ]
        ]);
    } else {
        jsonResponse(['message' => 'Email atau password salah'], 401);
    }
}

// MIDDLEWARE: CEK TOKEN (SEMUA ROUTE DI BAWAH INI BUTUH LOGIN)

$headers = getallheaders();
$jwt = null;
if (isset($headers['Authorization'])) {
    $matches = [];
    if (preg_match('/Bearer\s(\S+)/', $headers['Authorization'], $matches)) {
        $jwt = $matches[1];
    }
}

if (!$jwt) {
    jsonResponse(['message' => 'Token tidak ditemukan. Silakan Login.'], 401);
}

try {
    $decoded = JWT::decode($jwt, new Key($secret_key, $algo));
    $userData = $decoded->data; 
} catch (Exception $e) {
    jsonResponse(['message' => 'Token tidak valid atau kadaluarsa'], 401);
}

// ROUTE PRIVATE (BUTUH TOKEN)

//  GET USER PROFILE
if ($method === 'GET' && $path === '/profile') {
    // Ambil data user terbaru dari DB berdasarkan ID dari token
    $stmt = $db->prepare("SELECT id, name, email, role, created_at FROM users WHERE id = ?");
    $stmt->execute([$userData->id]);
    $profile = $stmt->fetch(PDO::FETCH_ASSOC);

    if (!$profile) {
        jsonResponse(['message' => 'User tidak ditemukan'], 404);
    }

    jsonResponse([
        'message' => 'Data Profil',
        'data' => $profile
    ]);
}

// --- ENDPOINT: UPDATE PROFILE (Edit Nama, Email, Password) ---
if (($method === 'PUT' || $method === 'PATCH') && $path === '/profile') {
    $input = json_decode(file_get_contents("php://input"));
    $userId = $userData->id;

    // 1. Ambil data lama dulu
    $stmt = $db->prepare("SELECT name, email, password FROM users WHERE id = ?");
    $stmt->execute([$userId]);
    $currentUser = $stmt->fetch(PDO::FETCH_ASSOC);

    if (!$currentUser) {
        jsonResponse(['message' => 'User tidak ditemukan'], 404);
    }

    // 2. Tentukan data baru (Pakai data input, kalau kosong pakai data lama)
    $newName = isset($input->name) && !empty($input->name) ? $input->name : $currentUser['name'];
    $newEmail = isset($input->email) && !empty($input->email) ? $input->email : $currentUser['email'];
    
    // Logic Password: Kalau input password diisi, update. Kalau kosong, pakai password lama.
    $newPasswordHash = $currentUser['password'];
    if (isset($input->password) && !empty($input->password)) {
        $newPasswordHash = password_hash($input->password, PASSWORD_BCRYPT);
    }

    // 3. Validasi: Jika Email berubah, cek apakah email baru sudah dipakai orang lain?
    if ($newEmail !== $currentUser['email']) {
        $checkStmt = $db->prepare("SELECT id FROM users WHERE email = ? AND id != ?");
        $checkStmt->execute([$newEmail, $userId]);
        if ($checkStmt->rowCount() > 0) {
            jsonResponse(['message' => 'Email sudah digunakan oleh user lain'], 409);
        }
    }

    // 4. Eksekusi Update
    try {
        $updateStmt = $db->prepare("UPDATE users SET name = ?, email = ?, password = ? WHERE id = ?");
        $updateStmt->execute([$newName, $newEmail, $newPasswordHash, $userId]);
        
        jsonResponse([
            'message' => 'Profil berhasil diperbarui',
            'data' => [
                'id' => $userId,
                'name' => $newName,
                'email' => $newEmail
            ]
        ]);
    } catch (Exception $e) {
        jsonResponse(['message' => 'Gagal mengupdate profil: ' . $e->getMessage()], 500);
    }
}

// --- ENDPOINT: ADMIN DASHBOARD (Hanya Admin) ---
if ($method === 'GET' && $path === '/admin-dashboard') {
    if ($userData->role !== 'admin') {
        jsonResponse(['message' => 'Akses Ditolak. Khusus Admin.'], 403);
    }

    jsonResponse([
        'message' => 'Selamat datang di Admin Dashboard',
        'secret_data' => 'Ini data rahasia hanya untuk admin.'
    ]);
}

jsonResponse(['message' => 'Endpoint Not Found'], 404);