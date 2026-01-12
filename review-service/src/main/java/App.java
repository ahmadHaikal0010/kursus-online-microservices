import io.javalin.Javalin;
import com.mongodb.client.*;
import org.bson.Document;
import org.bson.types.ObjectId;
import java.util.*;
import static com.mongodb.client.model.Filters.eq;

public class App {

    public static void main(String[] args) {

        String mongoUri = System.getenv().getOrDefault(
                "MONGO_URI",
                "mongodb://admin:admin123@mongo-db:27017/reviewdb?authSource=admin");

        MongoClient client = MongoClients.create(mongoUri);
        MongoDatabase db = client.getDatabase("reviewdb");
        MongoCollection<Document> reviews = db.getCollection("reviews");

        Javalin app = Javalin.create().start(5002);

        // POST /reviews
        app.post("/reviews", ctx -> {
            Document body = Document.parse(ctx.body());
            reviews.insertOne(body);
            ctx.json(body);
        });

        // GET /reviews
        app.get("/reviews", ctx -> {
            List<Map<String, Object>> result = new ArrayList<>();

            for (Document doc : reviews.find()) {
                Map<String, Object> item = new HashMap<>();

                item.put("_id", doc.getObjectId("_id").toHexString()); // 🔥 INI
                item.put("product_id", doc.getInteger("product_id"));
                item.put("review", doc.getString("review"));
                item.put("rating", doc.getInteger("rating"));

                result.add(item);
            }

            ctx.json(result);
        });

        // GET /reviews/product/{id}
        app.get("/reviews/product/{id}", ctx -> {
            int productId = Integer.parseInt(ctx.pathParam("id"));
            List<Map<String, Object>> result = new ArrayList<>();

            for (Document doc : reviews.find(eq("product_id", productId))) {
                Map<String, Object> item = new HashMap<>();

                item.put("_id", doc.getObjectId("_id").toHexString()); // 🔥 INI
                item.put("product_id", doc.getInteger("product_id"));
                item.put("review", doc.getString("review"));
                item.put("rating", doc.getInteger("rating"));

                result.add(item);
            }

            ctx.json(result);
        });

        // PUT /reviews/{id}
        app.put("/reviews/{id}", ctx -> {
            String id = ctx.pathParam("id");

            Document body = Document.parse(ctx.body());

            Document update = new Document("$set", new Document()
                    .append("review", body.getString("review"))
                    .append("rating", body.getInteger("rating")));

            var result = reviews.updateOne(
                    eq("_id", new ObjectId(id)),
                    update);

            if (result.getMatchedCount() == 0) {
                ctx.status(404).json(Map.of("message", "Review not found"));
            } else {
                ctx.json(Map.of(
                        "message", "Review updated successfully",
                        "id", id));
            }
        });

        // DELETE /reviews/{id}
        app.delete("/reviews/{id}", ctx -> {
            String id = ctx.pathParam("id");

            var result = reviews.deleteOne(eq("_id", new ObjectId(id)));

            if (result.getDeletedCount() == 0) {
                ctx.status(404).json(Map.of("message", "Review not found"));
            } else {
                ctx.json(Map.of(
                        "message", "Review deleted successfully",
                        "id", id));
            }
        });

    }
}
