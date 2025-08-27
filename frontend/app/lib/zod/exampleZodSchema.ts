import { z } from "zod";

/**
 * Example user schema for demonstration purposes.
 * Validates a user object with name, email, and age fields.
 */
export const exampleUserSchema = z.object({
  name: z
    .string()
    .min(2, { message: "Name must be at least 2 characters long." })
    .max(50, { message: "Name must be at most 50 characters long." }),
  email: z.string().email({ message: "Invalid email address." }),
  age: z
    .number()
    .int({ message: "Age must be an integer." })
    .min(0, { message: "Age must be a positive number." })
    .max(120, { message: "Age must be less than or equal to 120." }),
});

// Example usage:
// const result = exampleUserSchema.safeParse({ name: "Alice", email: "alice@example.com", age: 30 });
// if (!result.success) {
//   console.error(result.error);
// }
