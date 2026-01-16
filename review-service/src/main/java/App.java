import io.javalin.Javalin;
import java.sql.*;
import java.util.*;

public class App {

    private static final String DB_URL = "jdbc:sqlite:reviews.db"; // relative path

    public static void main(String[] args) throws Exception {

        // Koneksi database
        Connection conn = DriverManager.getConnection(DB_URL);

        // Shutdown hook untuk menutup koneksi
        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            try {
                conn.close();
                System.out.println("SQLite connection closed");
            } catch (Exception e) {
                e.printStackTrace();
            }
        }));

        // CREATE TABLE jika belum ada
        try (Statement stmt = conn.createStatement()) {
            stmt.execute("""
                CREATE TABLE IF NOT EXISTS reviews (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    product_id INTEGER NOT NULL,
                    review TEXT NOT NULL,
                    rating INTEGER NOT NULL
                )
            """);
        }

        // Start Javalin
        Javalin app = Javalin.create().start(5002);

        // POST /reviews
        app.post("/reviews", ctx -> {
            ReviewRequest body = ctx.bodyAsClass(ReviewRequest.class);

            try (PreparedStatement ps = conn.prepareStatement(
                    "INSERT INTO reviews (product_id, review, rating) VALUES (?, ?, ?)",
                    Statement.RETURN_GENERATED_KEYS)) {

                ps.setInt(1, body.product_id);
                ps.setString(2, body.review);
                ps.setInt(3, body.rating);

                int inserted = ps.executeUpdate(); // **Wajib executeUpdate**
                try (ResultSet rs = ps.getGeneratedKeys()) {
                    rs.next();
                    ctx.json(Map.of(
                            "id", rs.getInt(1),
                            "message", "Review created"
                    ));
                }
            }
        });

        // GET /reviews
        app.get("/reviews", ctx -> {
            List<Map<String, Object>> list = new ArrayList<>();
            try (Statement stmt = conn.createStatement();
                 ResultSet rs = stmt.executeQuery("SELECT * FROM reviews")) {

                while (rs.next()) {
                    list.add(Map.of(
                            "id", rs.getInt("id"),
                            "product_id", rs.getInt("product_id"),
                            "review", rs.getString("review"),
                            "rating", rs.getInt("rating")
                    ));
                }
            }
            ctx.json(list);
        });

        // GET /reviews/product/{id}
        app.get("/reviews/product/{id}", ctx -> {
            int productId = Integer.parseInt(ctx.pathParam("id"));
            List<Map<String, Object>> list = new ArrayList<>();

            try (PreparedStatement ps = conn.prepareStatement(
                    "SELECT * FROM reviews WHERE product_id = ?")) {
                ps.setInt(1, productId);

                try (ResultSet rs = ps.executeQuery()) {
                    while (rs.next()) {
                        list.add(Map.of(
                                "id", rs.getInt("id"),
                                "product_id", rs.getInt("product_id"),
                                "review", rs.getString("review"),
                                "rating", rs.getInt("rating")
                        ));
                    }
                }
            }
            ctx.json(list);
        });

        // PUT /reviews/{id}
        app.put("/reviews/{id}", ctx -> {
            int id = Integer.parseInt(ctx.pathParam("id"));
            ReviewRequest body = ctx.bodyAsClass(ReviewRequest.class);

            try (PreparedStatement ps = conn.prepareStatement(
                    "UPDATE reviews SET review=?, rating=? WHERE id=?")) {
                ps.setString(1, body.review);
                ps.setInt(2, body.rating);
                ps.setInt(3, id);

                int updated = ps.executeUpdate();
                if (updated == 0) {
                    ctx.status(404).json(Map.of("message", "Review not found"));
                } else {
                    ctx.json(Map.of("message", "Review updated"));
                }
            }
        });

        // DELETE /reviews/{id}
        app.delete("/reviews/{id}", ctx -> {
            int id = Integer.parseInt(ctx.pathParam("id"));

            try (PreparedStatement ps = conn.prepareStatement(
                    "DELETE FROM reviews WHERE id=?")) {
                ps.setInt(1, id);

                int deleted = ps.executeUpdate();
                if (deleted == 0) {
                    ctx.status(404).json(Map.of("message", "Review not found"));
                } else {
                    ctx.json(Map.of("message", "Review deleted"));
                }
            }
        });

        System.out.println("Server started on http://localhost:5002");
    }

    // Class request untuk POST & PUT
    public static class ReviewRequest {
        public int product_id;
        public String review;
        public int rating;
    }
}
