package mongodbatlas

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	matlas "github.com/mongodb-partners/go-client-mongodb-atlas/mongodbatlas"
)

func TestAccResourceMongoDBAtlasDatabaseUser_basic(t *testing.T) {
	var dbUser matlas.DatabaseUser

	resourceName := "mongodbatlas_database_user.test"
	groupID := "5cf5a45a9ccf6400e60981b6" // Modify until project data source is created.

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMongoDBAtlasDatabaseUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasDatabaseUserConfig(groupID, "atlasAdmin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasDatabaseUserExists(resourceName, &dbUser),
					testAccCheckMongoDBAtlasDatabaseUserAttributes(&dbUser, "test-acc-username"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
					resource.TestCheckResourceAttr(resourceName, "username", "test-acc-username"),
					resource.TestCheckResourceAttr(resourceName, "password", "test-acc-password"),
					resource.TestCheckResourceAttr(resourceName, "database_name", "admin"),
					resource.TestCheckResourceAttr(resourceName, "roles.0.role_name", "atlasAdmin"),
				),
			},
			{
				Config: testAccMongoDBAtlasDatabaseUserConfig(groupID, "read"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasDatabaseUserExists(resourceName, &dbUser),
					testAccCheckMongoDBAtlasDatabaseUserAttributes(&dbUser, "test-acc-username"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
					resource.TestCheckResourceAttr(resourceName, "username", "test-acc-username"),
					resource.TestCheckResourceAttr(resourceName, "password", "test-acc-password"),
					resource.TestCheckResourceAttr(resourceName, "database_name", "admin"),
					resource.TestCheckResourceAttr(resourceName, "roles.0.role_name", "read"),
				),
			},
		},
	})

}

func testAccCheckMongoDBAtlasDatabaseUserExists(resourceName string, dbUser *matlas.DatabaseUser) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*matlas.Client)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		log.Printf("[DEBUG] groupId: %s", rs.Primary.Attributes["group_id"])

		if dbUserResp, _, err := conn.DatabaseUsers.Get(context.Background(), rs.Primary.Attributes["group_id"], rs.Primary.ID); err == nil {
			*dbUser = *dbUserResp
			return nil
		}

		return fmt.Errorf("database user(%s) does not exist", rs.Primary.ID)
	}
}

func testAccCheckMongoDBAtlasDatabaseUserAttributes(dbUser *matlas.DatabaseUser, username string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if dbUser.Username != username {
			return fmt.Errorf("bad username: %s", dbUser.Username)
		}

		return nil
	}
}

func testAccCheckMongoDBAtlasDatabaseUserDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*matlas.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mongodbatlas_database_user" {
			continue
		}

		// Try to find the database user
		_, _, err := conn.DatabaseUsers.Get(context.Background(), rs.Primary.Attributes["group_id"], rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("database user (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccMongoDBAtlasDatabaseUserConfig(groupID, roleName string) string {
	return fmt.Sprintf(`
resource "mongodbatlas_database_user" "test" {
	username      = "test-acc-username"
	password      = "test-acc-password"
	group_id      = "%s"
	database_name = "admin"
	
	roles {
		role_name     = "%s"
		database_name = "admin"
	}
}
`, groupID, roleName)
}