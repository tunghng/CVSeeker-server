# Generated by Django 5.0.6 on 2024-05-18 05:35

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('main', '0002_relevanceai_email_name_scrapin_email_name'),
    ]

    operations = [
        migrations.AddField(
            model_name='phantombuster',
            name='remain_time',
            field=models.IntegerField(default=0),
        ),
        migrations.AddField(
            model_name='providermanagement',
            name='number_errors',
            field=models.IntegerField(default=0),
        ),
    ]
